package menuController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MenuController")

type MenuController struct {
	config      *config.Config
	validate    *validator.Validate
	menuSvc     service.IMenuService
	userRoleSvc service.IUserRoleService
	roleSvc     service.IRoleService
}

func New(
	config *config.Config,
	validate *validator.Validate,
	menuSvc service.IMenuService,
	userRoleSvc service.IUserRoleService,
	roleSvc service.IRoleService,
) *MenuController {
	return &MenuController{
		config:      config,
		validate:    validate,
		menuSvc:     menuSvc,
		userRoleSvc: userRoleSvc,
		roleSvc:     roleSvc,
	}
}
