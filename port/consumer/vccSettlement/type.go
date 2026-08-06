package vccsettlement

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("VccSettlementConsumer")

type handler struct {
	logger  logger.ILogger
	service service.IVccSettlementService
}

func New(log logger.ILogger, service service.IVccSettlementService) consumer.IVccSettlementConsumer {
	return &handler{logger: log, service: service}
}
