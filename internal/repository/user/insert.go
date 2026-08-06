package user

import (
	"context"
	"errors"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Create a new user.
func (r *UserRepository) Create(ctx context.Context, user *userModel.User) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "users")

	query := `
		INSERT INTO users (
		                   uuid, name, email, status, password, blocked_at, merchant_id, last_login_at, is_change_password,
		                   deactivate_at, created_at, updated_at, deleted_at
		                   )
		VALUES (
		        :uuid, :name, :email, :status, :password, :blocked_at, :merchant_id, :last_login_at, :is_change_password,
		        :deactivate_at, :created_at, :updated_at, :deleted_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		r.logger.Error(ctx, "error when inserting user", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting user", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
