package payoutManualProcessingAccount

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMPayoutManualProcessingAccountController")

type CRMPayoutManualProcessingAccountController struct {
	service  service.IPayoutManualProcessingAccountService
	validate *validator.Validate
}

func New(
	service service.IPayoutManualProcessingAccountService,
	validate *validator.Validate,
) controller.V1CRMPayoutManualProcessingAccountController {
	return &CRMPayoutManualProcessingAccountController{
		service:  service,
		validate: validate,
	}
}
