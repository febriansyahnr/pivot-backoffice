package role

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMRoleController")

type CRMRoleController struct {
	roleSvc  service.IRoleService
	validate *validator.Validate
	logger   pdkLogger.ILogger
}

type CRMRoleControllerOptions func(*CRMRoleController)

func New(
	roleSvc service.IRoleService,
	validate *validator.Validate,
	opts ...CRMRoleControllerOptions) controller.V1CRMRoleController {
	c := &CRMRoleController{
		roleSvc:  roleSvc,
		validate: validate,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithLogger(l pdkLogger.ILogger) CRMRoleControllerOptions {
	return func(cc *CRMRoleController) {
		cc.logger = l
	}
}
