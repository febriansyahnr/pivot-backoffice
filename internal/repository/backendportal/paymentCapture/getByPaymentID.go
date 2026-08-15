package paymentCapture

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *paymentCaptureRepository) GetByPaymentID(ctx context.Context, paymentID string) ([]*paymentCaptureModel.PaymentCapture, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/paymentCapture/GetByPaymentID")
	defer span.End()

	query := `
		SELECT id, payment_id, processor_reference_id, status, 
		       release_remaining_amount, currency, amount, created_at, updated_at
		FROM payment_captures 
		WHERE payment_id = ?
		ORDER BY created_at DESC
	`

	var captures []*paymentCaptureModel.PaymentCapture
	err := r.db.SelectContext(ctx, &captures, query, paymentID)

	if err != nil {
		r.logger.Error(ctx, "Failed to get payment captures by payment ID", logger.Error(err), logger.String("paymentID", paymentID))
		return nil, err
	}

	return captures, nil
}
