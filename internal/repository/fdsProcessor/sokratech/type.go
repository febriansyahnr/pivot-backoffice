package sokratech

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SokratechRepository")

type repository struct {
	config config.SokratechConfig
	secret config.SokratechSecret
	client httpRequestExt.IHTTPRequest
	logger logger.ILogger
}

type WorkflowRequester interface {
	GetWorkflowID() string
	GetWorkflowName() string
	GetWorkflowPayload() any
}

func New(config config.SokratechConfig, secret config.SokratechSecret, client httpRequestExt.IHTTPRequest, log logger.ILogger) port.IWorkflowFDSRepository {
	return &repository{config, secret, client, log}
}
