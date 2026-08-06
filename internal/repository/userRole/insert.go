package userRole

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserRoleRepository) Create(ctx context.Context, ur *userRole.UserRole) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userRole/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_role")

	query := `
		INSERT INTO user_role (uuid, user_id, role_id, created_at, updated_at, deleted_at)
		VALUES (:uuid, :user_id, :role_id, :created_at, :updated_at, :deleted_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, ur)
	if err != nil {
		r.logger.Error(ctx, "error when inserting user_role", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting user_role", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
