package paymentCapture

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
)

func (r *paymentCaptureRepository) Insert(ctx context.Context, capture *paymentCaptureModel.PaymentCapture) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/paymentCapture/Insert")
	defer span.End()

	query := `
		INSERT INTO payment_captures (
			id, payment_id, processor_reference_id, status, 
			release_remaining_amount, currency, amount, created_at, updated_at
		) VALUES (
			:id, :payment_id, :processor_reference_id, :status, 
			:release_remaining_amount, :currency, :amount, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, capture)

	if err != nil {
		r.logger.Error(ctx, "Failed to insert payment capture", logger.Error(err))
		return err
	}

	return nil
}
