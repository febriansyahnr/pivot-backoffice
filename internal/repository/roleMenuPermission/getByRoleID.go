package roleMenuPermissionRepository

import (
	"context"

	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleMenuPermissionRepository) GetByRoleID(ctx context.Context, roleID string) ([]*roleMenuPermissionModel.RoleMenuPermission, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/roleMenuPermission/GetByRoleID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "role_menu_permission")

	query := `
		SELECT role_id, menu_id, permission_id
		FROM role_menu_permission
		WHERE role_id = ?
	`

	var permissions []*roleMenuPermissionModel.RoleMenuPermission
	if err := r.db.SelectContext(ctx, &permissions, query, roleID); err != nil {
		r.logger.Error(ctx, "error when getting role menu permissions", logger.Error(err))
		return nil, err
	}

	return permissions, nil
}
