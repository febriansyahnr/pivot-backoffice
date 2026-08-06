package roleMenuPermissionRepository

import (
	"context"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleMenuPermissionRepository) DeleteByMenuAndPermissions(ctx context.Context, roleID, menuID string, permissionIDs []string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/roleMenuPermission/DeleteByMenuAndPermissions")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "role_menu_permission")

	if len(permissionIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(permissionIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
		DELETE FROM role_menu_permission
		WHERE role_id = ? AND menu_id = ? AND permission_id IN (%s)
	`, placeholders)

	args := []interface{}{roleID, menuID}
	for _, permID := range permissionIDs {
		args = append(args, permID)
	}

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		r.logger.Error(ctx, "error when deleting role menu permissions", logger.Error(err))
		return err
	}

	return nil
}
