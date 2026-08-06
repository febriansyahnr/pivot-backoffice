package roleMenuPermissionService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RoleMenuPermissionService")

type RoleMenuPermissionService struct {
	repo   repository.IRoleMenuPermissionRepository
	logger logger.ILogger
}

func New(repo repository.IRoleMenuPermissionRepository, logger logger.ILogger) service.IRoleMenuPermissionService {
	return &RoleMenuPermissionService{
		repo:   repo,
		logger: logger,
	}
}
