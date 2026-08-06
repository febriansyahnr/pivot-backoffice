package refundRepository

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RefundRepository) Insert(ctx context.Context, refund *refundModel.Refund) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/refund/Insert")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	query := `
	INSERT INTO refunds (
		uuid, merchant_id, client_reference_id, payment_id, payment_charge_id,
		currency, amount, status, reason, description,
		destination_type, method, created_at, updated_at, metadata
	)
	VALUES (
		:uuid, :merchant_id, :client_reference_id, :payment_id, :payment_charge_id,
		:currency, :amount, :status, :reason, :description,
		:destination_type, :method, :created_at, :updated_at, :metadata
	)`

	affected, err := r.db.NamedExecContext(ctx, query, refund)
	if err != nil {
		r.logger.Error(ctx, "error when inserting refunds", logger.Error(err))
		return err
	}
	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting refunds, no affected rows", logger.Error(err))
		return err
	}

	return nil
}
