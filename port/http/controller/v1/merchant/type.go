package merchant

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantController")

type MerchantController struct {
	merchantSvc service.IMerchantService
	feeSvc      service.IFeeService
	productSvc  service.IProductService
	validate    *validator.Validate
	rabbitMqExt rabbitMqExt.IRabbitMQExt

	config *config.Config
}

type ControllerFunc func(*MerchantController)

func New(
	merchantSvc service.IMerchantService,
	validate *validator.Validate,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	depends ...ControllerFunc) *MerchantController {
	c := &MerchantController{
		merchantSvc: merchantSvc,
		validate:    validate,
		rabbitMqExt: rabbitMqExt,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithFeeService(svc service.IFeeService) ControllerFunc {
	return func(c *MerchantController) {
		c.feeSvc = svc
	}
}

func WithProductService(svc service.IProductService) ControllerFunc {
	return func(c *MerchantController) {
		c.productSvc = svc
	}
}

func WithConfig(config *config.Config) ControllerFunc {
	return func(c *MerchantController) {
		c.config = config
	}
}
