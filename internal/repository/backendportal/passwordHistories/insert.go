package passwordHistories

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Insert old password
func (r *PasswordHistoriesRepository) Insert(ctx context.Context, uuid string, userId string, password string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/passwordHistories/Insert")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "password_histories")

	query := `
		INSERT INTO password_histories (uuid, user_id, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`

	affected, err := r.db.ExecContext(ctx, query, uuid, userId, password, time.Now())
	if err != nil {
		r.logger.Error(ctx, "error when inserting password histories", logger.Error(err))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting password histories", logger.Error(err))
		return err
	}

	return nil
}
