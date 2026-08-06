package paymentCaptureConsumer

import (
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentCaptureConsumer")

type paymentCaptureConsumer struct {
	logger logger.ILogger

	unifiedPaymentSvc services.IUnifiedPaymentService
}

func New(
	logger logger.ILogger,
	unifiedPaymentSvc services.IUnifiedPaymentService,
) consumer.IPaymentCaptureConsumer {
	return &paymentCaptureConsumer{
		logger:            logger,
		unifiedPaymentSvc: unifiedPaymentSvc,
	}
}