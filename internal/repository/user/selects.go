package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

const (
	SelectUserStr = `u.uuid, u.name, u.email, u.status, u.password, u.blocked_at, u.last_login_at,
		u.deactivate_at, u.merchant_id, COALESCE(m.name, '') as merchant_name, u.refresh_token, u.is_change_password,
		r.name as 'role', r.uuid as 'role_id', u.pin_hash, u.created_at, u.updated_at, u.deleted_at, m.status as merchant_status, m.reason_status,
		IFNULL(u.email = m.pic_email, FALSE) AS as_merchant_pic, u.totp_status, COALESCE(u.preferred_2fa_method, '') as preferred_2fa_method`
)

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/ListUsers")
	defer segment.End()

	var users []*userModel.User

	query := `
		SELECT ` + SelectUserStr + ` FROM
			users as u
		LEFT JOIN
			user_role as ur 
		ON
		    u.uuid = ur.user_id
		LEFT JOIN
			roles as r
		ON
			ur.role_id = r.uuid
		LEFT JOIN
			merchants as m
		ON
			u.merchant_id = m.uuid
		LIMIT ? OFFSET ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	if err := r.db.SelectContext(ctx, &users, query, limit, offset); err != nil {
		r.logger.Error(ctx, "error when listing users", logger.Error(err))
		return users, err
	}

	return users, nil
}

func (r *UserRepository) ListUsersByMerchantID(
	ctx context.Context,
	filter *userModel.ListUsersByMerchantIDRequest,
	page, perPage int64) (*commonModel.PaginationResponse, error) {

	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/UserRepository/ListUsersByMerchantID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*userModel.User, 0)
		errG       = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT 
    	` + SelectUserStr + `
	FROM users AS u
	LEFT JOIN
		user_role as ur 
	ON
		u.uuid = ur.user_id
	LEFT JOIN
		roles as r
	ON
		ur.role_id = r.uuid
	LEFT JOIN
		merchants as m
	ON
		u.merchant_id = m.uuid`

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "u.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "u.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "u.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}
	if filter.Name != "" {
		conditions = append(conditions, "u.name LIKE ?")
		args = append(args, "%"+filter.Name+"%")
	}
	if filter.RoleID != "" {
		conditions = append(conditions, "r.uuid = ?")
		args = append(args, filter.RoleID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	sortOrder := "DESC"
	sortCol := "u.created_at"
	if filter.SortBy == constant.UserSortColName {
		sortCol = "u.name"
	}

	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}

	querySort := fmt.Sprintf(" ORDER BY %s %s", sortCol, sortOrder)
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get user list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := `SELECT COUNT(u.uuid) as totalItems
		FROM users AS u
		LEFT JOIN user_role as ur ON u.uuid = ur.user_id
		LEFT JOIN roles as r ON ur.role_id = r.uuid
		LEFT JOIN merchants as m ON u.merchant_id = m.uuid`

	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
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

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}

func (r *UserRepository) FindUserByID(ctx context.Context, id string) (*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/FindUserByID")
	defer segment.End()

	var user userModel.User

	query := `
		SELECT ` + SelectUserStr + ` FROM
			users as u
		LEFT JOIN 
			user_role as ur 
		ON
		    u.uuid = ur.user_id
		LEFT JOIN
			roles as r
		ON
			ur.role_id = r.uuid
		LEFT JOIN
			merchants as m
		ON
			u.merchant_id = m.uuid
		WHERE u.uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "user not found", logger.String("id", id))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding user", logger.Error(err))
		return &user, err
	}

	return &user, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/FindUserByEmail")
	defer segment.End()

	var user userModel.User

	query := `
		SELECT ` + SelectUserStr + ` FROM
			users as u
		LEFT JOIN
			user_role as ur 
		ON
		    u.uuid = ur.user_id
		LEFT JOIN
			roles as r
		ON
			ur.role_id = r.uuid
		LEFT JOIN
			merchants as m
		ON
			u.merchant_id = m.uuid
		WHERE u.email = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "user not found", logger.String("email", email))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding user by email", logger.Error(err))
		return &user, err
	}

	return &user, nil
}

func (r *UserRepository) FindUserTOTPDataByID(ctx context.Context, userId string) (user *userModel.UserTOTPData, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/user/FindUserTOTPDataByID")
	defer segment.End()

	user, ctx = &userModel.UserTOTPData{}, context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
		uuid, email, totp_wrapped_secret, totp_encrypt_version, totp_status, updated_at
	FROM users
	WHERE uuid = ?;`

	if err = r.db.GetContext(ctx, user, rawQuery, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
