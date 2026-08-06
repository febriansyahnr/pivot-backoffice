package role

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"

	"github.com/go-playground/validator/v10"
)

var otelTracer = otel.Tracer("RoleController")

type RoleController struct {
	roleSvc       service.IRoleService
	permissionSvc service.IPermissionService
	validate      *validator.Validate
}

func New(
	validate *validator.Validate,
	roleSvc service.IRoleService,
	permissionSvc service.IPermissionService,
) *RoleController {
	return &RoleController{
		validate:      validate,
		roleSvc:       roleSvc,
		permissionSvc: permissionSvc,
	}
}
