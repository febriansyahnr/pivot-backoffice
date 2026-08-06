package v1CrmPaymentController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1CrmPaymentController")

type handler struct {
	validator  *validatorExt.Validate
	paymentSvc service.IPaymentService
}

func New(paymentSvc service.IPaymentService) controller.V1CRMPaymentController {
	return &handler{
		paymentSvc: paymentSvc,
		validator:  validatorExt.New(),
	}
}
