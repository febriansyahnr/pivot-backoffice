package paymentRepository

import (
	"context"
	"database/sql"
	"errors"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *PaymentRepository) FindPendingSubsequentCardFundedPayout(ctx context.Context, merchantID, referenceID string) ([]model.CardFundedPayment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/FindPendingSubsequentCardFundedPayout")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,customers")

	rawQuery := `SELECT
		p.uuid, p.merchant_id, p.reference_id, p.currency, p.amount, p.fee, 
		p.metadata->>'$.cardFundedPayout.count' AS count,
		p.metadata->>'$.cardFundedPayout.sequence' AS sequence,
		p.metadata->>'$.cardFundedPayout.firstPaymentId' AS first_payment_id, 
		c.metadata->>'$.paymentMethods[0].card.fingerprint' AS card_fingerprint, p.created_at,
		at.uuid AS charge_id
	FROM
		payments p
	JOIN customers c ON c.uuid = p.metadata->>'$.cardFundedPayout.cardId'
	JOIN account_transactions at ON at.reference_id = p.uuid AND at.type = 'PAYMENT' AND at.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
	WHERE
		p.merchant_id = ? AND p.reference_id = ?
		AND p.status = 'REQUIRE_ACTION' 
		AND p.metadata->>'$.cardFundedPayout.sequence' > 1
		AND p.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY);`

	result := []model.CardFundedPayment{}
	if err := r.db.SelectContext(ctx, &result, rawQuery, merchantID, referenceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *PaymentRepository) GetCardFundedPayoutFundingSummary(ctx context.Context, merchantID, referenceID string, maxCreatedDays int) (*model.CardFundedPayoutFundingSummary, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetCardFundedPayoutFundingSummary")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions")

	rawQuery := `SELECT
		p.reference_id AS payout_id,
		p.merchant_id AS merchant_id,
		SUM(IF(at.type = 'PAYMENT', at.credit, 0)) AS total_payment,
		SUM(IF(at.type = 'PAYMENT' AND at.status = 'PENDING', at.credit, 0)) AS total_waiting,
		SUM(IF(at.type = 'PAYMENT' AND at.status = 'FAILED', at.credit, 0)) AS total_failed,
		SUM(IF(at.type = 'PAYMENT' AND at.status = 'SUCCESS' AND IFNULL(at.settlement_status, 'PENDING') = 'PENDING', at.credit, 0)) AS total_pending_settlement,
		SUM(IF(at.type = 'PAYMENT' AND at.status = 'SUCCESS' AND IFNULL(at.settlement_status, 'PENDING') = 'SUCCESS', at.credit, 0)) AS total_success_settlement,
		SUM(IF(at.type = 'FEE' AND at.status = 'SUCCESS', at.debit, 0)) AS total_fee
	FROM payments p
	JOIN account_transactions at ON at.reference_id = p.uuid AND at.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
	WHERE 
		p.merchant_id = ? AND p.reference_id = ? AND p.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
	GROUP BY merchant_id, payout_id;`

	result := model.CardFundedPayoutFundingSummary{}
	if err := r.db.GetContext(ctx, &result, rawQuery, maxCreatedDays, merchantID, referenceID, maxCreatedDays); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *PaymentRepository) HardDeleteCardFundedPayoutPayments(ctx context.Context, merchantID, referenceID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/HardDeleteCardFundedPayoutPayments")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions")

	rawQuery := `DELETE p, at FROM payments p JOIN account_transactions at ON at.reference_id = p.uuid WHERE p.merchant_id = ? AND p.reference_id = ?;`

	if _, err := r.db.ExecContext(ctx, rawQuery, merchantID, referenceID); err != nil {
		return err
	}
	return nil
}
