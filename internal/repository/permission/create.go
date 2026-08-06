package permissionRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PermissionRepository) Create(ctx context.Context, permission *permissionModel.Permission) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/permission/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "permissions")

	query := "INSERT INTO permissions (uuid, slug, name, description, `group`, created_at, updated_at) " +
		"VALUES (:uuid, :slug, :name, :description, :group, :created_at, :updated_at)"

	affected, err := r.db.NamedExecContext(ctx, query, permission)
	if err != nil {
		r.logger.Error(ctx, "error when inserting permission", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting permission", logger.Error(err))
		return err
	}

	return nil
}
