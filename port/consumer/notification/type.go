package notification

import (
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("NotificationConsumer")

type handler struct {
	logger logger.ILogger
	rmq    rabbitMqExt.IRabbitMQExt
}

func New(logger logger.ILogger, rmq rabbitMqExt.IRabbitMQExt) consumer.INotificationConsumer {
	return &handler{logger, rmq}
}
