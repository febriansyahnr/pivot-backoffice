package settlementConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SettlementConsumer")

type consumerHandler struct {
	logger        logger.ILogger
	settlementSvc service.ISettlementService
}

func New(logger logger.ILogger, settlementSvc service.ISettlementService) consumer.ISettlementConsumer {
	return &consumerHandler{logger, settlementSvc}
}
