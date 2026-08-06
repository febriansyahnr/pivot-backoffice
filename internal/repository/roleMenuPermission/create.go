package roleMenuPermissionRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleMenuPermissionRepository) Create(ctx context.Context, pivot *roleMenuPermissionModel.RoleMenuPermission) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/roleMenuPermission/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "role_menu_permission")

	query := "INSERT INTO role_menu_permission (role_id, menu_id, permission_id) " +
		"VALUES (:role_id, :menu_id, :permission_id)"

	affected, err := r.db.NamedExecContext(ctx, query, pivot)
	if err != nil {
		r.logger.Error(ctx, "error when inserting role menu permission", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting role menu permission", logger.Error(err))
		return err
	}

	return nil
}
