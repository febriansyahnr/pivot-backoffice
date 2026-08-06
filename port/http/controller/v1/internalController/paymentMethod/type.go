package internalPaymentMethodController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalPaymentMethodController")

type InternalPaymentMethodController struct {
	MerchantTopUpSvc service.IMerchantTopUpService
	PaymentMethodSvc service.IPaymentMethodService
}

func New(
	merchantTopUpSvc service.IMerchantTopUpService,
	paymentMethodSvc service.IPaymentMethodService,
) controller.V1InternalPaymentMethodController {
	return &InternalPaymentMethodController{
		MerchantTopUpSvc: merchantTopUpSvc,
		PaymentMethodSvc: paymentMethodSvc,
	}
}
