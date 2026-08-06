package xbCoreProcessorRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("XbCoreProcessorRepository")

type xbCoreProcessorRepository struct {
	config      *config.Config
	secret      *config.Secret
	logger      logger.ILogger
	httpRequest httpRequestExt.IHTTPRequest
}

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpRequest httpRequestExt.IHTTPRequest,
) repository.IXbCoreProcessorRepository {
	return &xbCoreProcessorRepository{
		config:      config,
		secret:      secret,
		logger:      logger,
		httpRequest: httpRequest,
	}
}
