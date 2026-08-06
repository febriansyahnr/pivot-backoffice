package notificationService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("NotificationService")

type notificationService struct {
	config      *config.Config
	logger      logger.ILogger
	rabbitMqExt rabbitMqExt.IRabbitMQExt
}

func New(config *config.Config, logger logger.ILogger, rmq rabbitMqExt.IRabbitMQExt) service.INotificationService {
	return &notificationService{
		config:      config,
		logger:      logger,
		rabbitMqExt: rmq,
	}
}
