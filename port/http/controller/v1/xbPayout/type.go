package xbPayoutController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("xbPayoutController")

type xbPayoutController struct {
	config   *config.Config
	secret   *config.Secret
	validate *validator.Validate

	xbPayoutSvc     service.IXbPayoutService
	merchantSvc     service.IMerchantService
	disbursementSvc service.IDisbursementService
}

type ControllerFunc func(*xbPayoutController)

func New(
	config *config.Config,

	depends ...ControllerFunc,
) controller.V1XbPayoutController {
	c := &xbPayoutController{
		config:   config,
		validate: validatorExt.New(),
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithXbPayoutService(svc service.IXbPayoutService) ControllerFunc {
	return func(c *xbPayoutController) {
		c.xbPayoutSvc = svc
	}
}

func WithMerchantService(svc service.IMerchantService) ControllerFunc {
	return func(c *xbPayoutController) {
		c.merchantSvc = svc
	}
}

func WithDisbursementService(svc service.IDisbursementService) ControllerFunc {
	return func(c *xbPayoutController) {
		c.disbursementSvc = svc
	}
}

func WithSecret(secret *config.Secret) ControllerFunc {
	return func(c *xbPayoutController) {
		c.secret = secret
	}
}
