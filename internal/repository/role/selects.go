package role

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleRepository) GetList(ctx context.Context, filter *role.GetRoleFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*role.Role, 0)
		errG       = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT 
	r.uuid AS uuid,
	r.name AS name,
	r.type AS type,
	r.slug AS slug,
	r.created_at as created_at,
	r.updated_at as updated_at,
	r.deleted_at as deleted_at,
	GROUP_CONCAT(DISTINCT p.group ORDER BY p.group) AS permissions
	FROM roles AS r
	LEFT JOIN role_menu_permission AS rp ON r.uuid = rp.role_id
	LEFT JOIN permissions AS p ON rp.permission_id = p.uuid WHERE r.deleted_at IS NULL`

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "(r.merchant_id = ? OR r.merchant_id IS NULL)")
		args = append(args, filter.MerchantID)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Query builder
	queryGroupBy := " GROUP BY r.uuid, r.name, r.type"
	querySort := " ORDER BY r.type DESC, r.created_at ASC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += queryGroupBy + querySort + queryLimitOffset

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get role list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(uuid) as totalItems FROM roles r WHERE r.deleted_at IS NULL"
	if len(conditions) > 0 {
		queryCount += " AND  " + strings.Join(conditions, " AND ")
	}

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalItems, queryCount, args...)
		if err != nil {
			mu.Lock()
			totalItems = 0
			mu.Unlock()
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))
	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	validData := make([]*role.RoleResponse, len(data))
	for i, r := range data {
		validData[i] = r.ToResponse()
	}

	return &commonModel.PaginationResponse{
		Data: validData,
		Meta: meta,
	}, nil
}

func (r *RoleRepository) FindRoleByID(ctx context.Context, id string) (*role.Role, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/FindRoleByID")
	defer segment.End()

	var roles role.Role

	query := `
		SELECT
			r.uuid, r.merchant_id, r.name, r.slug, r.type, r.created_at, r.updated_at, r.deleted_at, GROUP_CONCAT(DISTINCT p.group ORDER BY p.group) AS permissions
		FROM roles AS r
		LEFT JOIN role_menu_permission AS rp ON r.uuid = rp.role_id
		LEFT JOIN permissions AS p ON rp.permission_id = p.uuid
		WHERE r.uuid = ? AND r.deleted_at IS NULL
		GROUP BY r.uuid, r.name, r.type`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	if err := r.db.GetContext(ctx, &roles, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "role not found", logger.String("id", id))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding role", logger.Error(err))
		return &roles, err
	}

	return &roles, nil
}

func (r *RoleRepository) FindRoleBySlug(ctx context.Context, slug string) (*role.Role, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/FindRoleBySlug")
	defer segment.End()

	var roles role.Role

	query := `
		SELECT
			uuid, merchant_id, name, slug, type, created_at, updated_at, deleted_at
		FROM roles
		WHERE slug = ?`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	if err := r.db.GetContext(ctx, &roles, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "role not found", logger.String("slug", slug))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding role", logger.Error(err))
		return &roles, err
	}

	return &roles, nil
}

func (r *RoleRepository) TotalRoleByMerchantID(ctx context.Context, merchantID string) (total uint64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/role/TotalRoleByMerchantID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	rawQuery := `SELECT COUNT(uuid) FROM roles WHERE merchant_id = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, &total, rawQuery, merchantID); err != nil {
		r.logger.Error(ctx, "db.GetContext: Total role by merchant id", logger.Error(err))
	}
	return
}

func (r *RoleRepository) CheckAvailableRoleName(ctx context.Context, merchantID, roleName string) (avail bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/role/CheckAvailableRoleName")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	rawQuery := `SELECT IFNULL(COUNT(uuid), 0) = 0  FROM roles WHERE merchant_id = ? AND name = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, &avail, rawQuery, merchantID, roleName); err != nil {
		r.logger.Error(ctx, "db.GetContext: Check available role name", logger.Error(err))
	}
	return
}
