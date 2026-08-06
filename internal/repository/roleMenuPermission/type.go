package roleMenuPermissionRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RoleMenuPermissionRepository")

type RoleMenuPermissionRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IRoleMenuPermissionRepository {
	return &RoleMenuPermissionRepository{
		db:     db,
		logger: logger,
	}
}
