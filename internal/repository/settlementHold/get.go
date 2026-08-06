package settlementHold

import (
	"context"
	"database/sql"
	"errors"

	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *settlementHoldRepo) GetByPaymentID(ctx context.Context, paymentId string) (*settlementHold.SettlementHold, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/settlementHold/GetByPaymentID")
	defer segment.End()

	var settlementHold settlementHold.SettlementHold
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	query := `
		SELECT uuid, merchant_id, payment_id, status, created_by, created_at, updated_at, deleted_at
		FROM ` + tableName + `
		WHERE payment_id = ? AND deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &settlementHold, query, paymentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when retrieve settlement hold data", logger.Error(err), logger.Any("paymentId", paymentId))
		return nil, err
	}

	return &settlementHold, nil
}
