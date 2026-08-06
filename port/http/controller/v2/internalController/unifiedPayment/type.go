package v2InternalUnifiedPaymentController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V2InternalUnifiedPaymentController")

type paymentController struct {
	config   *config.Config
	validate *validatorExt.Validate
	monitor  *monitoring.Monitor
	logger   logger.ILogger

	unifiedPaymentSvc services.IUnifiedPaymentService
	customerSvc       services.ICustomerService
}

type ControllerFunc func(*paymentController)

func New(
	config *config.Config,
	monitor *monitoring.Monitor,
	options ...ControllerFunc,
) controller.V2InternalUnifiedPaymentController {
	c := &paymentController{
		config:   config,
		validate: validatorExt.New(),
		monitor:  monitor,
	}
	for _, fn := range options {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *paymentController) {
		c.logger = log
	}
}

func WithUnifiedPaymentService(svc services.IUnifiedPaymentService) ControllerFunc {
	return func(c *paymentController) {
		c.unifiedPaymentSvc = svc
	}
}

func WithCustomerService(svc services.ICustomerService) ControllerFunc {
	return func(c *paymentController) {
		c.customerSvc = svc
	}
}
