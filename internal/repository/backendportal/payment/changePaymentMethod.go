package paymentRepository

import (
	"context"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *PaymentRepository) ChangePaymentMethod(ctx context.Context, payment *paymentModel.PaymentDTO) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/ChangePaymentMethod")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	rawQuery := `UPDATE
			payments
		SET
			payment_method_id = :payment_method_id, processor_reference_number = :processor_reference_number, 
			metadata = :metadata, payment_url = :payment_url, expired_at = :expired_at, updated_at = :updated_at,
			customer_id = :customer_id, currency = :currency, amount = :amount, fee = :fee, discount = :discount, total_amount = :total_amount
		WHERE uuid = :uuid;`

	_, err := r.db.NamedExecContext(ctx, rawQuery, payment)
	return err
}
