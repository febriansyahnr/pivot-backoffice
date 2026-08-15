package callbackRepository

import (
	"context"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
)

const TableCallbackEvents = "callback_events"

func (r *CallbackRepository) GetCallbackEvents(ctx context.Context) ([]callbackModel.CallbackEvent, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetCallbackEvents")
	defer segment.End()

	var events []callbackModel.CallbackEvent

	query := `
		SELECT
			uuid,
			event,
			label,
			event_group,
			is_active,
			created_at,
			updated_at
		FROM ` + TableCallbackEvents + `
		WHERE is_active = 1
		ORDER BY event_group, event`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, TableCallbackEvents)

	if err := r.db.SelectContext(ctx, &events, query); err != nil {
		r.logger.Error(ctx, "error when getting callback events", logger.Error(err))
		return nil, err
	}

	return events, nil
}
