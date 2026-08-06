package settlementHold

import (
	"context"

	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *settlementHoldRepo) Update(ctx context.Context, data *settlementHold.SettlementHold, history *settlementHold.SettlementHoldHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/settlementHold/Update")
	defer segment.End()

	ctx, err := r.db.BeginTxx(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when start db transaction", logger.Error(err))
		return err
	}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	query := `
		UPDATE ` + tableName + `
		SET status = ? , updated_at = ?
		WHERE uuid = ?
	`
	_, err = r.db.ExecContext(ctx, query, data.Status, data.UpdatedAt, data.UUID)
	if err != nil {
		r.logger.Error(ctx, "error when updating settlement hold data", logger.Error(err), logger.Any("payload", data))
		r.rollbackTrx(ctx)
		return err
	}

	err = r.CreateHistory(ctx, history)
	if err != nil {
		r.rollbackTrx(ctx)
		return err
	}

	err = r.db.Commit(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when commit db transaction", logger.Error(err))
		r.rollbackTrx(ctx)
		return err
	}

	return nil
}
