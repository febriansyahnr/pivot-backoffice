package callbackRepository

import (
	"context"
	"time"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CallbackRepository) UpdateCallbackLog(ctx context.Context, callbackLog *callbackModel.CallbackLog) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/UpdateCallbackLog")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "callback_logs")

	query := `UPDATE callback_logs
		SET
			event = :event, response = :response, status = :status, retry = :retry, reference_id = :reference_id, updated_at = :updated_at
		WHERE uuid = :uuid
	`

	_, err := r.db.NamedExecContext(ctx, query, callbackLog)
	if err != nil {
		r.logger.Error(ctx, "error when updating callback logs", logger.Error(err))
		return err
	}

	return nil
}

func (r *CallbackRepository) UpdateCallbackURLById(ctx context.Context, id, url string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/callback/UpdateCallbackURLById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableCallback)
	rawQuery := "UPDATE " + TableCallback + " SET url = ?, updated_at = ? WHERE uuid = ?"

	_, err := r.db.ExecContext(ctx, rawQuery, url, time.Now().UTC(), id)
	return err
}

func (r *CallbackRepository) UpdateCallbackBaseURLById(ctx context.Context, id, url string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/callback/UpdateCallbackBaseURLById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableCallback)
	rawQuery := "UPDATE " + TableCallback + " SET base_url = ?, updated_at = ? WHERE uuid = ?"

	_, err := r.db.ExecContext(ctx, rawQuery, url, time.Now().UTC(), id)
	return err
}
