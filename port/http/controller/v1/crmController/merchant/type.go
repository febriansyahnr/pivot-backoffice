package merchant

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMMerchantController")

type CRMMerchantController struct {
	merchantSvc service.IMerchantService
	userSvc     service.IUserService
	tncSvc      service.ITNCService
	validate    *validator.Validate
	rabbitMqExt rabbitMqExt.IRabbitMQExt

	config *config.Config
	logger pdkLogger.ILogger
}

type CRMMerchantControllerOptions func(*CRMMerchantController)

func New(
	merchantSvc service.IMerchantService,
	userSvc service.IUserService,
	validate *validator.Validate,
	rabbitMqExt rabbitMqExt.IRabbitMQExt, opts ...CRMMerchantControllerOptions) controller.V1CRMMerchantController {
	c := &CRMMerchantController{
		merchantSvc: merchantSvc,
		userSvc:     userSvc,
		validate:    validate,
		rabbitMqExt: rabbitMqExt,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithConfig(config *config.Config) CRMMerchantControllerOptions {
	return func(c *CRMMerchantController) {
		c.config = config
	}
}

func WithLogger(l pdkLogger.ILogger) CRMMerchantControllerOptions {
	return func(cc *CRMMerchantController) {
		cc.logger = l
	}
}

func WithTNCService(tncSvc service.ITNCService) CRMMerchantControllerOptions {
	return func(cc *CRMMerchantController) {
		cc.tncSvc = tncSvc
	}
}
