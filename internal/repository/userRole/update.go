package userRole

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserRoleRepository) UpdateByUserID(ctx context.Context, ur *userRole.UserRole) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userRole/UpdateByUserID")
	defer segment.End()

	oldUserId := ur.UserID

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_role")

	query := `
			UPDATE
				user_role
			SET user_id = ?, role_id = ?, updated_at = ?
			WHERE
			    user_id = ?`
	_, err := r.db.ExecContext(ctx, query, ur.UserID, ur.RoleID, ur.UpdatedAt, oldUserId)
	if err != nil {
		r.logger.Error(ctx, "error when updating user_role by user_id", logger.Error(err))
		return err
	}

	return nil
}
