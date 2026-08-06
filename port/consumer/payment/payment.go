package paymentConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentConsumer")

type paymentConsumer struct {
	paymentSvc        service.IPaymentService
	merchantTopUpSvc  service.IMerchantTopUpService
	orchestratorSvc   service.IOrchestratorService
	rabbitMqExt       rabbitMqExt.IRabbitMQExt
	merchantSvc       service.IMerchantService
	unifiedPaymentSvc service.IUnifiedPaymentService
	conf              *config.Config
	logger            logger.ILogger
}

func New(
	conf *config.Config,
	logger logger.ILogger,
	paymentSvc service.IPaymentService,
	merchantTopUpSvc service.IMerchantTopUpService,
	orchestratorSvc service.IOrchestratorService,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	merchantSvc service.IMerchantService,
	unifiedPaymentSvc service.IUnifiedPaymentService,
) consumer.PaymentConsumer {
	return &paymentConsumer{
		conf:              conf,
		logger:            logger,
		paymentSvc:        paymentSvc,
		merchantTopUpSvc:  merchantTopUpSvc,
		orchestratorSvc:   orchestratorSvc,
		rabbitMqExt:       rabbitMqExt,
		merchantSvc:       merchantSvc,
		unifiedPaymentSvc: unifiedPaymentSvc,
	}
}
