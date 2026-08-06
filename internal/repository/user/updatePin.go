package user

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserRepository) UpdatePin(ctx context.Context, userID, hashedPin string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/UpdatePin")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "users")

	query := `UPDATE users SET pin_hash = ? where uuid = ?`
	_, err := r.db.ExecContext(ctx, query, hashedPin, userID)
	if err != nil {
		r.logger.Error(ctx, "error when updating user pin", logger.Error(err))
		return err
	}

	return nil
}
