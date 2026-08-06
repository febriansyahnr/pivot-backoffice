package charges

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ChargesController")

type ChargesController struct {
	config   *config.Config
	validate *validator.Validate
	monitor  *monitoring.Monitor

	unifiedPaymentService service.IUnifiedPaymentService
	merchantService       service.IMerchantService

	logger logger.ILogger
}

type ControllerFunc func(*ChargesController)

func New(
	config *config.Config,
	validate *validator.Validate,
	monitor *monitoring.Monitor,
	depends ...ControllerFunc,
) controller.V1ChargesController {
	c := &ChargesController{
		config:   config,
		validate: validate,
		monitor:  monitor,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *ChargesController) {
		c.logger = log
	}
}

func WithUnifiedPaymentService(svc service.IUnifiedPaymentService) ControllerFunc {
	return func(c *ChargesController) {
		c.unifiedPaymentService = svc
	}
}

func WithMerchantService(svc service.IMerchantService) ControllerFunc {
	return func(c *ChargesController) {
		c.merchantService = svc
	}
}