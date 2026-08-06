package internalXbController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("internalXbController")

type InternalXbController struct {
	config   *config.Config
	secret   *config.Secret
	validate *validator.Validate
	logger   logger.ILogger

	xbPayoutSvc     service.IXbPayoutService
	merchantSvc     service.IMerchantService
	disbursementSvc service.IDisbursementService
}

type ControllerFunc func(*InternalXbController)

func New(
	config *config.Config,
	depends ...ControllerFunc,
) controller.V1InternalXbController {
	c := &InternalXbController{
		config:   config,
		validate: validatorExt.New(),
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithXbPayoutService(svc service.IXbPayoutService) ControllerFunc {
	return func(c *InternalXbController) {
		c.xbPayoutSvc = svc
	}
}

func WithMerchantService(svc service.IMerchantService) ControllerFunc {
	return func(c *InternalXbController) {
		c.merchantSvc = svc
	}
}

func WithDisbursementService(svc service.IDisbursementService) ControllerFunc {
	return func(c *InternalXbController) {
		c.disbursementSvc = svc
	}
}

func WithSecret(secret *config.Secret) ControllerFunc {
	return func(c *InternalXbController) {
		c.secret = secret
	}
}

func WithLogger(logger logger.ILogger) ControllerFunc {
	return func(c *InternalXbController) {
		c.logger = logger
	}
}
