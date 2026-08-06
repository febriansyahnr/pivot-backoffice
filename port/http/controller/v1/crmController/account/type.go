package account

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMAccountController")

type handler struct {
	validator       *validatorExt.Validate
	accountSvc      service.IAccountService
	orchestratorSvc service.IOrchestratorService
}

func New(
	accountSvc service.IAccountService,
	orchestratorSvc service.IOrchestratorService,
) controller.V1CRMAccountController {
	return &handler{
		validator:       validatorExt.New(),
		accountSvc:      accountSvc,
		orchestratorSvc: orchestratorSvc,
	}
}
