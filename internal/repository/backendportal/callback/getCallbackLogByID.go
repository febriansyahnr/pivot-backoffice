package callbackRepository

import (
	"context"
	"database/sql"
	"errors"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CallbackRepository) GetCallbackLogByID(ctx context.Context, id string) (*callbackModel.CallbackLogWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetCallbackLogByID")
	defer segment.End()

	var callbackLog callbackModel.CallbackLogWithMaster

	query := `
		SELECT 
			cl.uuid, c.merchant_id, cl.callback_id, cm.name as 'type', cl.event, c.base_url, c.url,
			cl.request, cl.response, cl.status, cl.retry, cl.reference_id, cl.created_at, cl.updated_at
		FROM callback_logs cl
		LEFT JOIN callbacks c ON cl.callback_id = c.uuid
		LEFT JOIN callback_masters cm ON c.callback_master_id = cm.uuid
		WHERE cl.uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, TableCallbackLog)

	if err := r.db.GetContext(ctx, &callbackLog, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "callback not found", logger.String("id", id))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding callback", logger.Error(err))
		return &callbackLog, err
	}

	return &callbackLog, nil
}
