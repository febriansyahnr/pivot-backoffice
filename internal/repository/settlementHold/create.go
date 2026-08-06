package settlementHold

import (
	"context"

	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *settlementHoldRepo) Create(ctx context.Context, data *settlementHold.SettlementHold, history *settlementHold.SettlementHoldHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/settlementHold/Create")
	defer segment.End()

	ctx, err := r.db.BeginTxx(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when start db transaction", logger.Error(err))
		return err
	}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	query := `
		INSERT INTO ` + tableName + `
			(uuid, merchant_id, payment_id, status, created_by, created_at, updated_at, deleted_at)
		VALUES (:uuid, :merchant_id, :payment_id, :status, :created_by, :created_at, :updated_at, :deleted_at)
	`
	_, err = r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when inserting settlement hold data", logger.Error(err), logger.Any("payload", data))
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

func (r *settlementHoldRepo) CreateHistory(ctx context.Context, history *settlementHold.SettlementHoldHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/settlementHold/CreateHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, historyTableName)
	query := `
		INSERT INTO ` + historyTableName + `
			(uuid, settlement_hold_id, status, reason, created_by, created_at, deleted_at)
		VALUES (:uuid, :settlement_hold_id, :status, :reason, :created_by, :created_at, :deleted_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, history)
	if err != nil {
		r.logger.Error(ctx, "error when inserting settlement hold history", logger.Error(err), logger.Any("payload", history))
		return err
	}

	return nil
}

func (r *settlementHoldRepo) rollbackTrx(ctx context.Context) {
	err := r.db.Rollback(ctx)
	if err != nil {
		r.logger.Error(ctx, "error rollback db transaction", logger.Error(err))
	}

}
