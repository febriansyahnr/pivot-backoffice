package permissionRepository

import (
	"context"
	"database/sql"
	"errors"

	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PermissionRepository) FindBySlug(ctx context.Context, slug string) (*permissionModel.Permission, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/permission/FindBySlug")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "permissions")

	var permission permissionModel.Permission

	query := `
		SELECT
			uuid, slug, name, description, 'group', created_at, updated_at
		FROM permissions
		WHERE slug = ?`

	if err := r.db.GetContext(ctx, &permission, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "permission not found", logger.String("slug", slug))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding permission", logger.Error(err))
		return nil, err
	}

	return &permission, nil
}

func (r *PermissionRepository) FindByRoleId(ctx context.Context, roleId string) ([]*permissionModel.Permission, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/permission/FindByRoleId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "permissions")

	var permission []*permissionModel.Permission

	query := `
		SELECT
			p.uuid, p.slug, p.name, p.description, p.group, p.created_at, p.updated_at
		FROM permissions as p
		LEFT JOIN role_menu_permission AS rp ON p.uuid = rp.permission_id
		LEFT JOIN roles AS r ON rp.role_id = r.uuid
		WHERE rp.role_id = ?`

	if err := r.db.SelectContext(ctx, &permission, query, roleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "failed to get permissions by role_id, not found", logger.String("role_id", roleId))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding permissions by role_id", logger.Error(err))
		return nil, err
	}

	return permission, nil
}
