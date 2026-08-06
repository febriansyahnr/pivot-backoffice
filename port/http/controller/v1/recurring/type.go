package recurring

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RecurringContractController")

type RecurringContractController struct {
	config   *config.Config
	validate *validator.Validate
	monitor  *monitoring.Monitor

	recurringContractService service.IRecurringContractService
	merchantService          service.IMerchantService

	logger logger.ILogger
}

type ControllerFunc func(*RecurringContractController)

func New(
	config *config.Config,
	validate *validator.Validate,
	monitor *monitoring.Monitor,
	depends ...ControllerFunc,
) controller.V1RecurringContractController {
	c := &RecurringContractController{
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
	return func(c *RecurringContractController) {
		c.logger = log
	}
}

func WithRecurringContractService(svc service.IRecurringContractService) ControllerFunc {
	return func(c *RecurringContractController) {
		c.recurringContractService = svc
	}
}

func WithMerchantService(svc service.IMerchantService) ControllerFunc {
	return func(c *RecurringContractController) {
		c.merchantService = svc
	}
}
