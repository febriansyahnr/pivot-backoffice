package payment

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentController")

type PaymentController struct {
	config   *config.Config
	validate *validator.Validate
	monitor  *monitoring.Monitor

	paymentService        service.IPaymentService
	unifiedPaymentService service.IUnifiedPaymentService
	merchantService       service.IMerchantService
	paymentMethodService  service.IPaymentMethodService
	userService           service.IUserService

	logger logger.ILogger
}

type ControllerFunc func(*PaymentController)

func New(
	config *config.Config,
	validate *validator.Validate,
	monitor *monitoring.Monitor,
	depends ...ControllerFunc,
) controller.V1PaymentController {
	c := &PaymentController{
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
	return func(c *PaymentController) {
		c.logger = log
	}
}

func WithPaymentService(svc service.IPaymentService) ControllerFunc {
	return func(c *PaymentController) {
		c.paymentService = svc
	}
}

func WithMerchantService(svc service.IMerchantService) ControllerFunc {
	return func(c *PaymentController) {
		c.merchantService = svc
	}
}

func WithPaymentMethodService(svc service.IPaymentMethodService) ControllerFunc {
	return func(c *PaymentController) {
		c.paymentMethodService = svc
	}
}

func WithUserService(svc service.IUserService) ControllerFunc {
	return func(c *PaymentController) {
		c.userService = svc
	}
}

func WithUnifiedPaymentService(svc service.IUnifiedPaymentService) ControllerFunc {
	return func(c *PaymentController) {
		c.unifiedPaymentService = svc
	}
}
