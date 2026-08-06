package commService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CommServiceHandler")

type handler struct {
	log     logger.ILogger
	service service.ICommService
}

func New(log logger.ILogger, service service.ICommService) consumer.CommServiceHandler {
	return &handler{log, service}
}
