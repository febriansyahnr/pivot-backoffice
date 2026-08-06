package beneficiaryAccountController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("BeneficiaryAccountController")

type Controller struct {
	config                *config.Config
	validate              *validator.Validate
	rabbitMqExt           rabbitMqExt.IRabbitMQExt
	beneficiaryAccountSvc service.IBeneficiaryAccountService
	disbursementSvc       service.IDisbursementService
	merchantSvc           service.IMerchantService
	feeSvc                service.IFeeService
}

func New(
	config *config.Config,
	validate *validator.Validate,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	beneficiaryAccountSvc service.IBeneficiaryAccountService,
	disbursementSvc service.IDisbursementService,
	merchantSvc service.IMerchantService,
	feeSvc service.IFeeService,
) controller.V1BeneficiaryAccountController {
	return &Controller{
		config:                config,
		validate:              validate,
		rabbitMqExt:           rabbitMqExt,
		beneficiaryAccountSvc: beneficiaryAccountSvc,
		disbursementSvc:       disbursementSvc,
		merchantSvc:           merchantSvc,
		feeSvc:                feeSvc,
	}
}
