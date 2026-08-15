package passwordHistories

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/passwordHistories"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *PasswordHistoriesRepository) FindByUserID(
	ctx context.Context,
	userId string,
	limit *int) ([]*passwordHistories.PasswordHistories, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/passwordHistories/FindByUserID")
	defer segment.End()

	var (
		histories []*passwordHistories.PasswordHistories
	)

	query := `
		SELECT
			uuid, user_id, password_hash, created_at
		FROM password_histories
		WHERE user_id = ? ORDER BY created_at DESC`

	if limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *limit)
	}

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "password_histories")

	// use SelectContext to execute a query that returns rows, typically a SELECT.
	err := r.db.SelectContext(ctx, &histories, query, userId)
	if err != nil {
		r.logger.Error(ctx, "error when find password histories by user id", logger.Error(err))
		return nil, err
	}

	return histories, nil
}

// FindByPassHashAndUserID is a function to find password histories by password hash and user id
func (r *PasswordHistoriesRepository) FindByPassHashAndUserID(
	ctx context.Context,
	userId string,
	passHash string) (*passwordHistories.PasswordHistories, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/passwordHistories/FindByPassHashAndUserID")
	defer segment.End()

	var (
		history passwordHistories.PasswordHistories
	)

	query := `
		SELECT
			uuid, user_id, password_hash, created_at
		FROM password_histories
		WHERE user_id = ? AND password_hash = ?`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "password_histories")

	// use GetContext to execute a query that returns a single row.
	err := r.db.GetContext(ctx, &history, query, userId, passHash)
	if err != nil {
		r.logger.Error(ctx, "error when find password histories by password hash and user id", logger.Error(err))
		return nil, err
	}

	return &history, nil
}
