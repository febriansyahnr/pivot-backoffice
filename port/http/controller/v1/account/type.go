package accountController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/tracer"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
)

var otelTracer = tracer.New("AccountController")

type account struct {
	config          *config.Config
	accountSvc      service.IAccountService
	orchestratorSvc service.IOrchestratorService
}

func New(
	config *config.Config,
	accountSvc service.IAccountService,
	orchestratorSvc service.IOrchestratorService,
) controller.V1AccountController {
	return &account{
		config:          config,
		accountSvc:      accountSvc,
		orchestratorSvc: orchestratorSvc,
	}
}
