package permissionRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PermissionRepository) Update(ctx context.Context, permission *permissionModel.Permission) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/permission/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "permissions")

	query := `UPDATE permissions p
			SET p.slug = ?, p.name = ?, p.description = ?, p.group = ?, p.updated_at = ? 
			WHERE p.uuid = ?`

	affected, err := r.db.ExecContext(ctx, query, permission.Slug, permission.Name, permission.Description, permission.Group, permission.UpdatedAt, permission.UUID)
	if err != nil {
		r.logger.Error(ctx, "error when updating permission", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when updating permission", logger.Error(err))
		return err
	}

	return nil
}
