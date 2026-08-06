package internalMerchantAuthController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalMerchantAuthController")

type InternalMerchantAuthController struct {
	validate    *validator.Validate
	merchantSvc service.IMerchantService
	logger      logger.ILogger
}

type OptionFunc func(*InternalMerchantAuthController)

func WithLogger(log logger.ILogger) OptionFunc {
	return func(h *InternalMerchantAuthController) {
		h.logger = log
	}
}

func New(validate *validator.Validate, merchantSvc service.IMerchantService, opts ...OptionFunc) controller.V1InternalMerchantAuthController {
	handler := &InternalMerchantAuthController{
		validate:    validate,
		merchantSvc: merchantSvc,
	}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}
