package roleAndPermissionCronHandler

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RoleAndPermissionCronHandler")

type RoleAndPermission struct {
	config                *config.Config
	logger                logger.ILogger
	roleSvc               service.IRoleService
	permissionSvc         service.IPermissionService
	menuSvc               service.IMenuService
	roleMenuPermissionSvc service.IRoleMenuPermissionService
	redisExt              redisExt.IRedisExt
}

type Option func(*RoleAndPermission)

func New(
	config *config.Config,
	logger logger.ILogger,
	roleSvc service.IRoleService,
	permissionSvc service.IPermissionService,
	menuSvc service.IMenuService,
	roleMenuPermissionSvc service.IRoleMenuPermissionService,
	options ...Option,
) handler.IRoleAndPermissionHandler {
	svc := &RoleAndPermission{
		config:                config,
		logger:                logger,
		roleSvc:               roleSvc,
		permissionSvc:         permissionSvc,
		menuSvc:               menuSvc,
		roleMenuPermissionSvc: roleMenuPermissionSvc,
	}

	for _, opt := range options {
		opt(svc)
	}
	return svc
}

func WithRedisClient(rdb redisExt.IRedisExt) Option {
	return func(rp *RoleAndPermission) {
		rp.redisExt = rdb
	}
}
