package conductor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	clientSdk "github.com/conductor-sdk/conductor-go/sdk/client"
	"github.com/conductor-sdk/conductor-go/sdk/worker"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"go.opentelemetry.io/otel/trace"
)

type workers struct {
	apiClient      *clientSdk.APIClient
	taskRunner     *TaskRunner
	taskNames      []string
	inProgressTask int64
}

type WorkerDefinition struct {
	TaskName string
	Handler  func(context.Context, *Task) (*TaskResult, error)
	// Number of tasks to be fetched in a single request.
	// Set to -1 to disable the worker, or 0 to use the default value.
	BatchSize    int
	PollInterval time.Duration
	PollTimeout  time.Duration
	Domain       string
}

// Blocking Function
func (w *workers) RunWorkers(ctx context.Context, workerDefs []WorkerDefinition) (err error) {
	if w.taskRunner == nil {
		return errors.New("task runner not initialized")
	}
	defer func() {
		if err == nil {
			return
		}
		for _, taskName := range w.taskNames {
			w.taskRunner.Shutdown(taskName)
		}
		w.taskNames = []string{}
	}()

	workers := make([]worker.Worker, 0, len(workerDefs))

	for _, def := range workerDefs {
		// Workers will not be registered if the batch size is set to -1.
		// If it's 0, the default value will be used.
		if def.BatchSize < 0 {
			continue
		}
		// Definition of worker and task performed
		options := []worker.Option{
			worker.WithBaseContext(ctx),
			worker.WithBatchSize(def.BatchSize),
			worker.WithPollInterval(def.PollInterval),
			worker.WithPollTimeout(def.PollTimeout),
		}
		if def.Domain != "" {
			options = append(options, worker.WithDomain(def.Domain))
		}
		w.taskNames = append(w.taskNames, def.TaskName)

		workers = append(workers, worker.NewWorker(def.TaskName, w.wrapHandler(ctx, def.Handler), options...))
	}

	if err = w.taskRunner.RegisterWorkers(workers...); err != nil {
		return fmt.Errorf("register workers: %w", err)
	}

	w.taskRunner.WaitWorkers()
	return nil
}

func (w *workers) wrapHandler(_ context.Context, handler func(context.Context, *Task) (*TaskResult, error)) func(t *Task) (any, error) {
	return func(t *Task) (_ any, err error) {
		atomic.AddInt64(&w.inProgressTask, 1)
		defer func() { atomic.AddInt64(&w.inProgressTask, -1) }()

		ctx, span := otelTracer.Start(context.Background(), "Task "+t.TaskDefName, trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		if t.ExternalInputPayloadStoragePath != "" {
			if t.InputData, err = w.getExternalPayload(ctx, t.ExternalInputPayloadStoragePath); err != nil {
				return NewTaskResultWithNonRetryableError(t, err), nil
			}
		}

		ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, t.TaskId)

		return handler(ctx, t)
	}
}

func (w *workers) Close(ctx context.Context) error {
	if w.taskRunner == nil {
		return nil
	}

	done := make(chan struct{}, 1)

	go func() {
		for _, taskName := range w.taskNames {
			w.taskRunner.SetBatchSize(taskName, 0)
		}

		for atomic.LoadInt64(&w.inProgressTask) > 0 {
		}
		time.Sleep(5 * time.Second)

		for _, taskName := range w.taskNames {
			w.taskRunner.Shutdown(taskName)

			log.Println("Shutdown Task:", taskName)
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// Currently only support external payload using PostgreSQL.
func (w *workers) getExternalPayload(ctx context.Context, payloadPath string) (result map[string]any, err error) {
	if resp, err := w.apiClient.Get(ctx, "/external/postgres/"+payloadPath, nil, &result); err != nil {
		return nil, fmt.Errorf("get external payload: %w", err)

	} else if resp == nil {
		return nil, errors.New("get external payload: http client response is nil")
	}
	return result, nil
}
