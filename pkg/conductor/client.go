package conductor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	clientSdk "github.com/conductor-sdk/conductor-go/sdk/client"
	logSdk "github.com/conductor-sdk/conductor-go/sdk/log"
	"github.com/conductor-sdk/conductor-go/sdk/settings"
	"github.com/conductor-sdk/conductor-go/sdk/worker"
	"github.com/conductor-sdk/conductor-go/sdk/workflow/executor"
)

type client struct {
	apiClient    *clientSdk.APIClient
	httpSettings *settings.HttpSettings
	mx           *sync.Mutex
	runner       *worker.TaskRunner
	executor     *executor.WorkflowExecutor
	health       IConductorHealth
}

func NewClient(cfg Config) (*client, error) {
	basicHttpSettings := settings.HttpSettings{
		BaseUrl: cfg.BaseURL,
		Headers: map[string]string{
			"Content-Type":    mimeApplicationJson,
			"Accept":          mimeApplicationJson,
			"Accept-Encoding": "gzip",
		},
	}

	if cfg.Authentication != nil {
		switch auth := cfg.Authentication.(type) {
		default:
			// Default handling

		case *BasicAuthentication:
			basicHttpSettings.Headers["Authorization"] = "Basic " + auth.Encode()
		}
	}

	// API Settings
	apiHttpSettings := basicHttpSettings
	apiHttpSettings.BaseUrl += "/api"
	apiClient := clientSdk.NewAPIClient(nil, &apiHttpSettings)

	if cfg.Logger != nil {
		logSdk.SetLogger(cfg.Logger)

	} else {
		logSdk.SetLogger(logSdk.NewNop())
	}

	client := &client{
		apiClient:    apiClient,
		httpSettings: &basicHttpSettings,
		health: &clientSdk.HealthCheckResourceApiService{
			APIClient: clientSdk.NewAPIClient(nil, &basicHttpSettings),
		},
		mx: new(sync.Mutex),
	}
	return client, nil
}

func (c *client) GetAPIClient() *clientSdk.APIClient {
	return c.apiClient
}

func (c *client) TaskRunner() *TaskRunner {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.runner == nil {
		c.runner = worker.NewTaskRunnerWithApiClient(c.apiClient)
	}
	return c.runner
}

func (c *client) WorkflowExecutor() *WorkflowExecutor {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.executor == nil {
		c.executor = executor.NewWorkflowExecutor(c.apiClient)
	}
	return c.executor
}

func (c *client) Workers() *workers {
	return &workers{apiClient: c.apiClient, taskRunner: c.TaskRunner()}
}

func (c *client) HealthCheck(ctx context.Context) error {
	ctx, span := otelTracer.Start(ctx, "conductor.HealthCheck")
	defer span.End()

	if status, _, err := c.health.DoCheck(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)

	} else if !status.Healthy {
		return errors.New("conductor is in an unhealthy state")
	}
	return nil
}
