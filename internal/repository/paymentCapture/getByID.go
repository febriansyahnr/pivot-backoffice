package paymentCapture

import (
	"context"
	"database/sql"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *paymentCaptureRepository) GetByID(ctx context.Context, id string) (*paymentCaptureModel.PaymentCapture, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/paymentCapture/GetByID")
	defer span.End()

	query := `
		SELECT id, payment_id, processor_reference_id, status, 
		       release_remaining_amount, currency, amount, created_at, updated_at
		FROM payment_captures 
		WHERE id = ?
	`

	var capture paymentCaptureModel.PaymentCapture
	err := r.db.GetContext(ctx, &capture, query, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error(ctx, "Failed to get payment capture by ID", logger.Error(err), logger.String("id", id))
		return nil, err
	}

	return &capture, nil
}
