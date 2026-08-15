package paymentCapture

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
)

func (r *paymentCaptureRepository) Update(ctx context.Context, capture *paymentCaptureModel.PaymentCapture) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/paymentCapture/Update")
	defer span.End()

	query := `
		UPDATE payment_captures 
		SET processor_reference_id = :processor_reference_id,
		    status = :status,
		    release_remaining_amount = :release_remaining_amount,
		    currency = :currency,
		    amount = :amount,
		    updated_at = :updated_at
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, capture)

	if err != nil {
		r.logger.Error(ctx, "Failed to update payment capture", logger.Error(err))
		return err
	}

	return nil
}
