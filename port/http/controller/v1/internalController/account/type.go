package internalAccountController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalAccountController")

type handler struct {
	validator       *validatorExt.Validate
	accountSvc      service.IAccountService
	orchestratorSvc service.IOrchestratorService
	logger          logger.ILogger
}

func New(
	accountSvc service.IAccountService,
	orchestratorSvc service.IOrchestratorService,
) controller.V1InternalAccountController {
	return &handler{
		validator:       validatorExt.New(),
		accountSvc:      accountSvc,
		orchestratorSvc: orchestratorSvc,
	}
}

func WithLogger(h controller.V1InternalAccountController, logger logger.ILogger) {
	h.(*handler).logger = logger
}
