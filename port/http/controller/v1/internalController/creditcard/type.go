package creditcard

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalCreditCardController")

type Controller struct {
	config   *config.Config
	validate *validator.Validate

	logger logger.ILogger

	monitor *monitoring.Monitor

	merchantSvc      service.IMerchantService
	creditcardSvc    service.ICreditCardService
	orchestratorSvc  service.IOrchestratorService
	customerSvc      service.ICustomerService
	paymentMethodSvc service.IPaymentMethodService
	paymentSvc       service.IPaymentService
}

type Services struct {
	MerchantSvc      service.IMerchantService
	CreditcardSvc    service.ICreditCardService
	OrchestratorSvc  service.IOrchestratorService
	CustomerSvc      service.ICustomerService
	PaymentMethodSvc service.IPaymentMethodService
	PaymentSvc       service.IPaymentService
}

func New(
	config *config.Config,
	validate *validator.Validate,
	logger logger.ILogger,
	monitor *monitoring.Monitor,
	services Services,
) controller.V1CreditCardController {
	return &Controller{
		config:           config,
		validate:         validate,
		logger:           logger,
		monitor:          monitor,
		merchantSvc:      services.MerchantSvc,
		creditcardSvc:    services.CreditcardSvc,
		orchestratorSvc:  services.OrchestratorSvc,
		customerSvc:      services.CustomerSvc,
		paymentMethodSvc: services.PaymentMethodSvc,
		paymentSvc:       services.PaymentSvc,
	}
}
