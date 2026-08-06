package submerchant

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SubMerchantInternalController")

type SubMerchantInternalController struct {
	merchantSvc     service.IMerchantService
	accountSvc      service.IAccountService
	orchestratorSvc service.IOrchestratorService
	validate        *validator.Validate
}

func New(
	merchantSvc service.IMerchantService,
	accountSvc service.IAccountService,
	orchestratorSvc service.IOrchestratorService,
	validate *validator.Validate) *SubMerchantInternalController {
	return &SubMerchantInternalController{
		merchantSvc:     merchantSvc,
		accountSvc:      accountSvc,
		orchestratorSvc: orchestratorSvc,
		validate:        validate,
	}
}
