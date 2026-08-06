package subMerchant

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SubMerchantController")

type SubMerchantController struct {
	merchantSvc         service.IMerchantService
	paymentService      service.IPaymentService
	feeSvc              service.IFeeService
	accountSvc          service.IAccountService
	orchestratorSvc     service.IOrchestratorService
	forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService
	productSvc          service.IProductService
	disbursementSvc     service.IDisbursementService
	validate            *validator.Validate
	rabbitMqExt         rabbitMqExt.IRabbitMQExt
}

type ControllerFunc func(*SubMerchantController)

func New(
	merchantSvc service.IMerchantService,
	accountSvc service.IAccountService,
	orchestratorSvc service.IOrchestratorService,
	forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService,
	validate *validator.Validate,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	depends ...ControllerFunc) *SubMerchantController {
	c := &SubMerchantController{
		accountSvc:          accountSvc,
		orchestratorSvc:     orchestratorSvc,
		merchantSvc:         merchantSvc,
		forbiddenUsecaseSvc: forbiddenUsecaseSvc,
		validate:            validate,
		rabbitMqExt:         rabbitMqExt,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithFeeService(svc service.IFeeService) ControllerFunc {
	return func(c *SubMerchantController) {
		c.feeSvc = svc
	}
}

func WithProductService(svc service.IProductService) ControllerFunc {
	return func(c *SubMerchantController) {
		c.productSvc = svc
	}
}

func WithDisbursementService(svc service.IDisbursementService) ControllerFunc {
	return func(c *SubMerchantController) {
		c.disbursementSvc = svc
	}
}

func WithPaymentService(svc service.IPaymentService) ControllerFunc {
	return func(c *SubMerchantController) {
		c.paymentService = svc
	}
}
