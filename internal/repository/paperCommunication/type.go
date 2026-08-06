package paperCommunication

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaperCommunicationRepository")

type paperCommunicationRepository struct {
	config  *config.PaperCommunication
	log     logger.ILogger
	httpReq httpRequestExt.IHTTPRequest
}

func New(cfg *config.PaperCommunication, log logger.ILogger, httpReq httpRequestExt.IHTTPRequest) repository.IPaperCommunicationRepository {
	return &paperCommunicationRepository{cfg, log, httpReq}
}
