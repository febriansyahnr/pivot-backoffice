package qris

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("QrisConsumer")

type consumerHandler struct {
	logger  logger.ILogger
	service service.IQrisService
}

func New(logger logger.ILogger, service service.IQrisService) consumer.IProcessor {
	return &consumerHandler{logger, service}
}
