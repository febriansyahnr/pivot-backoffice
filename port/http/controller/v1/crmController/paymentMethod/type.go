package crmPaymentMethodController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMPaymentMethodController")

type handler struct {
	validator        *validatorExt.Validate
	paymentMethodSvc service.IPaymentMethodService
	merchantSvc      service.IMerchantService

	config *config.Config
}

type ICRMPaymentMethodControllerOptions func(*handler)

func New(paymentMethodSvc service.IPaymentMethodService, opts ...ICRMPaymentMethodControllerOptions) controller.V1CRMPaymentMethodController {
	h := &handler{
		validator:        validatorExt.New(),
		paymentMethodSvc: paymentMethodSvc,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func WithMerchantService(merchantSvc service.IMerchantService) ICRMPaymentMethodControllerOptions {
	return func(h *handler) {
		h.merchantSvc = merchantSvc
	}
}

func WithConfig(config *config.Config) ICRMPaymentMethodControllerOptions {
	return func(h *handler) {
		h.config = config
	}
}
