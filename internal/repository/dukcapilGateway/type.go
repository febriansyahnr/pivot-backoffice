package dukcapilgatewayrepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DukcapilGatewayRepository")

type DukcapilGatewayRepository struct {
	config      *config.Config
	secret      *config.Secret
	logger      logger.ILogger
	httpRequest httpRequestExt.IHTTPRequest
}

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpReq httpRequestExt.IHTTPRequest,
) repository.IDukcapilGatewayRepository {
	return &DukcapilGatewayRepository{
		config:      config,
		secret:      secret,
		logger:      logger,
		httpRequest: httpReq,
	}
}
