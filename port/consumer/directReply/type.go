package directreply

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DirectReplyConsumer")

type directReply struct {
	logger              logger.ILogger
	routingProcessorSvc service.IRoutingProcessorService
}

func New(
	log logger.ILogger,
	routingProcessorSvc service.IRoutingProcessorService,
) consumer.IDirectReplyConsumer {
	return &directReply{
		logger:              log,
		routingProcessorSvc: routingProcessorSvc,
	}
}
