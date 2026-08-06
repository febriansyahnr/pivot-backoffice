package xbPayoutConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("XbPayoutConsumer")

type consumerHandler struct {
	logger  logger.ILogger
	service service.IXbPayoutService
}

func New(logger logger.ILogger, service service.IXbPayoutService) consumer.IXbPayoutConsumer {
	return &consumerHandler{logger, service}
}
