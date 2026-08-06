package simulationController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SimulationController")

type Handler struct {
	validator        *validator.Validate
	paymentMethodSvc service.IPaymentMethodService
	paymentSvc       service.IPaymentService
}

type SimulationOptFunc func(*Handler)

func New(vld *validatorExt.Validate, depends ...SimulationOptFunc) controller.V1SimulationController {
	h := &Handler{
		validator: vld,
	}

	for _, fn := range depends {
		fn(h)
	}

	return h
}

func WithPaymentMethodService(svc service.IPaymentMethodService) SimulationOptFunc {
	return func(h *Handler) {
		h.paymentMethodSvc = svc
	}
}

func WithPaymentService(svc service.IPaymentService) SimulationOptFunc {
	return func(h *Handler) {
		h.paymentSvc = svc
	}
}
