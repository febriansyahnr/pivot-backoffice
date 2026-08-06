package refundConsumer

import (
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RefundConsumer")

type refundConsumer struct {
	logger logger.ILogger

	refundSvc          services.IRefundService
	refundProcessorSvc services.IRefundProcessorService
	paymentSvc         services.IPaymentService
	orchestratorSvc    services.IOrchestratorService
}

func New(
	logger logger.ILogger,
	refundSvc services.IRefundService,
	refundProcessorSvc services.IRefundProcessorService,
	paymentSvc services.IPaymentService,
	orchestratorSvc services.IOrchestratorService,
) consumer.IRefundConsumer {
	return &refundConsumer{
		logger:             logger,
		refundSvc:          refundSvc,
		refundProcessorSvc: refundProcessorSvc,
		paymentSvc:         paymentSvc,
		orchestratorSvc:    orchestratorSvc,
	}
}
