package callbackWorker

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/port/worker"

	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CallbackWorker")

type handler struct {
	logger     logger.ILogger
	jsonBinder conductor.InputBinder
	service    service.ICallbackService
}

func New(log logger.ILogger, service service.ICallbackService) worker.IMerchantCallbackHandler {
	return &handler{
		logger:     log,
		jsonBinder: &conductor.JSONBinder{},
		service:    service,
	}
}

func noop(_ context.Context, task *conductor.Task) (*conductor.TaskResult, error) {
	taskResult := model.NewTaskResultFromTask(task)
	taskResult.Status = model.CompletedTask
	return taskResult, nil
}
