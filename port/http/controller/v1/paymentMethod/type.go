package paymentMethodController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentMethodController")

type PaymentMethodController struct {
	paymentMethodSvc service.IPaymentMethodService
}

func New(paymentMethodSvc service.IPaymentMethodService) *PaymentMethodController {
	return &PaymentMethodController{
		paymentMethodSvc: paymentMethodSvc,
	}
}
