package paymentRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PaymentRepository) BeginTransaction(ctx context.Context) (context.Context, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/BeginTransaction")
	defer segment.End()

	ctxTx, err := r.db.BeginTxx(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when begin tx transaction", logger.Error(err))
		return nil, err
	}

	return ctxTx, nil
}

func (r *PaymentRepository) CommitTransaction(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/CommitTransaction")
	defer segment.End()

	err := r.db.Commit(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when commit transaction", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) RollbackTransaction(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/RollbackTransaction")
	defer segment.End()

	err := r.db.Rollback(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when rollback transaction", logger.Error(err))
		return err
	}

	return nil
}
