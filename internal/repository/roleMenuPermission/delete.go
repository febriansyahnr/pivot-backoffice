package roleMenuPermissionRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *RoleMenuPermissionRepository) Delete(ctx context.Context, roleID string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/roleMenuPermission/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "role_menu_permission")

	_, err = r.db.ExecContext(
		ctx, `DELETE FROM role_menu_permission WHERE role_id = ?;`, roleID,
	)
	return
}
