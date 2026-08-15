package paymentRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
)

func (r *PaymentRepository) CountActiveStaticPayment(ctx context.Context, merchantID, paymentMethodID string) (int, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/payments/CountActiveStaticPayment")
	defer span.End()

	var totalItems = 0

	queryCount := `
		SELECT COUNT(p.uuid) as totalItems
		FROM payments p 
		JOIN merchants m
		ON p.merchant_id = m.uuid
		LEFT JOIN merchants sm
		ON p.merchant_id = sm.uuid AND 
			sm.kyc_status = 'NOT_REQUIRED'
		WHERE (p.merchant_id = ? OR sm.parent_id = ?) AND 
			p.type = ? AND 
			p.status = ? AND 
			p.payment_method_id = ?
	`
	err := r.db.GetContext(ctx, &totalItems, queryCount, merchantID, merchantID,
		constant.UnifiedPaymentTypeMultiple, constant.UnifiedStaticPaymentStatusActive, paymentMethodID)
	if err != nil {
		r.logger.Error(ctx, "error when count active static payment", logger.Error(err))
		return 0, err
	}

	return totalItems, nil
}

func (r *PaymentRepository) GetFirstActiveStaticQrisByMerchant(ctx context.Context, merchantID, partnerReferenceNo string) (*paymentModel.Payment, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/payments/GetFirstActiveStaticQrisByMerchant")
	defer span.End()

	var dto paymentModel.PaymentWithPaymentMethodDTO

	queryGet := `
		SELECT p.uuid,
				p.reference_id,
				p.merchant_id,
				p.customer_id,
				p.payment_method_id,
				p.processor_reference_number,
				p.currency,
				p.amount,
				p.fee,
				p.discount,
				p.total_amount,
				p.status,
				p.type,
				p.metadata,
				p.payment_url,
				p.created_at,
				p.expired_at,
				p.updated_at,
				p.deleted_at,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer',
				pm.bank_name as 'payment_method_bank_name'
		FROM payments p
		LEFT JOIN merchants m
		ON p.merchant_id = m.uuid
		LEFT JOIN payment_methods pm
		ON p.payment_method_id = pm.uuid
		WHERE p.merchant_id = ? AND
			pm.type = ? AND
			p.type = ? AND
			p.status = ? `

	var args []interface{}
	args = append(args, merchantID, constant.ChannelQris, constant.UnifiedPaymentTypeMultiple, constant.StatusActive)

	if partnerReferenceNo != "" {
		queryGet += ` AND p.reference_id = ? `
		args = append(args, partnerReferenceNo)
	}

	queryGet += `
		ORDER BY p.created_at ASC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &dto, queryGet, args...)
	if err != nil {
		r.logger.Error(ctx, "error when get first active static QRIS by merchant", logger.Error(err))
		return nil, err
	}

	// Convert DTO to Payment using existing PaymentFromDTO method
	var payment paymentModel.Payment
	payment.PaymentFromPaymentWithPaymentMethodDTO(&dto)

	return &payment, nil
}
