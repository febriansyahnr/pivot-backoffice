package disbursementRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) BeginTransaction(ctx context.Context) (context.Context, error) {
	ctxTx, err := r.db.BeginTxx(ctx)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when begin tx transaction", logger.Error(err))
		return nil, err
	}

	return ctxTx, nil
}

func (r *DisbursementRepository) CommitTransaction(ctx context.Context) error {
	err := r.db.Commit(ctx)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when commit transaction", logger.Error(err))
		return err
	}

	return nil
}

func (r *DisbursementRepository) RollbackTransaction(ctx context.Context) error {
	err := r.db.Rollback(ctx)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when rollback transaction", logger.Error(err))
		return err
	}

	return nil
}
