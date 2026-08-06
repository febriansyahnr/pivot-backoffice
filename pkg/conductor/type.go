package conductor

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/conductor-sdk/conductor-go/sdk/log"
	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/conductor-sdk/conductor-go/sdk/worker"
	"github.com/conductor-sdk/conductor-go/sdk/workflow/executor"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ConductorWorkflow")

type (
	Task             = model.Task
	TaskResult       = model.TaskResult
	TaskRunner       = worker.TaskRunner
	WorkflowExecutor = executor.WorkflowExecutor
	InputBinder      = worker.InputBinder
	JSONBinder       = worker.JSONBinder
)

type Config struct {
	BaseURL        string
	Authentication Authentication
	Logger         ILogger
}

type BasicAuthentication struct {
	Username string
	Password string
}

func (b *BasicAuthentication) Encode() string {
	return base64.RawURLEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", b.Username, b.Password))
}

type Authentication interface {
	Encode() string
}

type IWorkflow interface {
	HealthCheck(ctx context.Context) error
}

type IConductorHealth interface {
	DoCheck(ctx context.Context) (model.HealthCheckStatus, *http.Response, error)
}

type ILogger log.Logger
