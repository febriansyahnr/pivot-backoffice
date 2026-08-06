package menuRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MenuRepository) FindBySlug(ctx context.Context, slug string) (*menuModel.Menu, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/menu/FindBySlug")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus")

	var menu menuModel.Menu

	query := "SELECT uuid, slug, name, type, icon, path, level, `order`, parent_id, created_at, updated_at FROM menus WHERE slug = ?"

	if err := r.db.GetContext(ctx, &menu, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "menu not found", logger.String("slug", slug))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding menu", logger.Error(err))
		return nil, err
	}

	return &menu, nil
}

func (r *MenuRepository) FindBySlugWithPermissions(ctx context.Context, slug string) (*menuModel.MenuResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/menu/FindBySlugWithPermissions")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus")

	var menu menuModel.Menu

	query := `
		SELECT
			m.uuid, m.slug, m.name, m.type, m.icon, m.path, m.level, m.` + "`order`" + `, m.parent_id, m.created_at, m.updated_at,
			(
				SELECT GROUP_CONCAT(DISTINCT CONCAT(p.group, ':', p.slug) ORDER BY p.slug)
				FROM permissions p
				WHERE p.slug LIKE CONCAT(m.slug, '.%%')
			) AS permissions
		FROM menus m
		WHERE m.slug = ?
	`

	if err := r.db.GetContext(ctx, &menu, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "menu not found", logger.String("slug", slug))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding menu with permissions", logger.Error(err))
		return nil, err
	}

	return menu.ToResponse(), nil
}

func (r *MenuRepository) GetAll(ctx context.Context, filter *menuModel.GetAllFilterRequest) ([]*menuModel.MenuResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/menu/GetAll")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus")

	var (
		conditions []string
		args       []interface{}
		menus      = make([]*menuModel.Menu, 0)
	)

	query := "SELECT " +
		" m.uuid, m.slug, m.name, m.type, m.icon, m.path, m.level, m.`order`, m.parent_id, m.allowed_products, m.created_at, m.updated_at, " +
		" GROUP_CONCAT(DISTINCT CONCAT(p.group, ':', p.slug) ORDER BY p.slug) AS permissions" +
		" FROM menus m" +
		" JOIN role_menu_permission rmp ON m.uuid = rmp.menu_id" +
		" LEFT JOIN permissions p ON rmp.permission_id = p.uuid"

	// Query condition builder
	if filter.RoleID != "" {
		conditions = append(conditions, "rmp.role_id = ?")
		args = append(args, filter.RoleID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query group by
	query += " GROUP BY m.uuid ORDER BY m.`level`, m.`order`"

	if err := r.db.SelectContext(ctx, &menus, query, args...); err != nil {
		r.logger.Error(ctx, "error when finding all menus", logger.Error(err))
		return nil, err
	}

	validData := make([]*menuModel.MenuResponse, len(menus))
	for i, menu := range menus {
		validData[i] = menu.ToResponse()
	}

	return validData, nil
}

func (r *MenuRepository) GetMenuAndPermissionIDs(ctx context.Context, slug string, permissionSlug ...string) (res *menuModel.MenuAndPermissionIDs, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/menu/GetMenuAndPermissionIDs")
	defer segment.End()

	res = &menuModel.MenuAndPermissionIDs{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus,permissions")

	rawQuery := fmt.Sprintf(`
		SELECT 
			menus.uuid AS menu_id, menus.name AS menu_name,
			IFNULL((
				SELECT 
					JSON_UNQUOTE(JSON_ARRAYAGG(JSON_OBJECT('id', p.uuid, 'name', p.name)))
				FROM permissions p
				WHERE p.slug IN (%s)
			), '') AS permissions
		FROM menus WHERE slug = ?`, "'"+strings.Join(permissionSlug, `', '`)+"'")
	if err := r.db.GetContext(ctx, res, rawQuery, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "db.SelectContext: Get menu and permissions", logger.Error(err))
		return nil, err
	}
	_ = json.Unmarshal([]byte(res.PermissionsStr), &res.Permissions)
	return
}
