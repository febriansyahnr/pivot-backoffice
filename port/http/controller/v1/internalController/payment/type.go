package internalPaymentController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"

	"github.com/go-playground/validator/v10"
)

var otelTracer = otel.Tracer("InternalPaymentController")

type InternalPaymentController struct {
	validate             *validator.Validate
	paymentSvc           service.IPaymentService
	unifiedPaymentSvc    service.IUnifiedPaymentService
	merchantSvc          service.IMerchantService
	rabbitMqExt          rabbitMqExt.IRabbitMQExt
	logger               logger.ILogger
	config               *config.Config
}

type OptionFunc func(*InternalPaymentController)

func WithLogger(log logger.ILogger) OptionFunc {
	return func(h *InternalPaymentController) {
		h.logger = log
	}
}

func WithUnifiedPaymentService(svc service.IUnifiedPaymentService) OptionFunc {
	return func(h *InternalPaymentController) {
		h.unifiedPaymentSvc = svc
	}
}

func WithConfig(cfg *config.Config) OptionFunc {
	return func(h *InternalPaymentController) {
		h.config = cfg
	}
}


func New(validate *validator.Validate, paymentSvc service.IPaymentService, merchantSvc service.IMerchantService, rabbitMqExt rabbitMqExt.IRabbitMQExt, opts ...OptionFunc) controller.V1InternalPaymentController {
	handler := &InternalPaymentController{
		validate:    validate,
		paymentSvc:  paymentSvc,
		merchantSvc: merchantSvc,
		rabbitMqExt: rabbitMqExt,
	}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}
