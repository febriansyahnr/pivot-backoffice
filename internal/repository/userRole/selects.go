package userRole

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserRoleRepository) FindUserRoleByUserID(ctx context.Context, userID string) (*userRole.UserRole, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userRole/FindUserRoleByUserID")
	defer segment.End()

	var roles userRole.UserRole

	query := `
		SELECT
			uuid, user_id, role_id, created_at, updated_at, deleted_at
		FROM user_role
		WHERE user_id = ?`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_role")

	if err := r.db.GetContext(ctx, &roles, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "user_role not found", logger.String("user_id", userID))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding user_role", logger.Error(err))
		return &roles, err
	}

	return &roles, nil
}

func (r *UserRoleRepository) TotalActiveUsersByRoleID(ctx context.Context, id string) (total uint64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/userRole/TotalActiveUsersByRoleID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_role")

	err = r.db.GetContext(
		ctx, &total,
		`SELECT COUNT(uuid) FROM user_role WHERE role_id = ? AND deleted_at IS NULL;`, id,
	)
	return
}
