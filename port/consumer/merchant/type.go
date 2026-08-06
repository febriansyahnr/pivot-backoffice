package merchantConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantConsumer")

type merchantConsumer struct {
	merchantSvc service.IMerchantService
	conf        *config.Config
	logger      logger.ILogger
}

func New(
	conf *config.Config,
	logger logger.ILogger,
	merchantSvc service.IMerchantService,
) consumer.IMerchantConsumer {
	return &merchantConsumer{
		conf:        conf,
		logger:      logger,
		merchantSvc: merchantSvc,
	}
}
