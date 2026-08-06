package refundRepository

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RefundRepository) UpdateData(ctx context.Context, refund *refundModel.Refund) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/refund/UpdateData")
	defer span.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	refund.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE refunds
		SET
			currency = ?, 
			amount = ?, 
			status = ?, 
			reason = ?, 
			description = ?,
			destination_type = ?, 
			method = ?,  
			updated_at = ?
		WHERE uuid = ?`

	_, err := r.db.ExecContext(ctx, query, refund.Currency, refund.Amount, refund.Status, refund.Reason, refund.Description, refund.DestinationType, refund.Method, refund.UpdatedAt, refund.UUID)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating refund", logger.Error(err))
		return err
	}

	return nil
}
