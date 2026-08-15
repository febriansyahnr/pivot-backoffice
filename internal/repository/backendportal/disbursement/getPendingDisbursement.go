package disbursementRepository

import (
	"context"
	"time"

	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
)

// GetPendingTransactionsBetweenTimeForInquiryTransaction retrieves pending disbursement transactions within a specified time range
// that need inquiry status check to the processor. This excludes transactions with reason_type CUT_OFF_TIME or WAITING_MANUAL_ACTION.
func (r *DisbursementRepository) GetPendingTransactionsBetweenTimeForInquiryTransaction(ctx context.Context, from, to time.Time) ([]*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetPendingTransactionsBetweenTimeForInquiryTransaction")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConstant.CtxSQLTableNameKey, "disbursements,account_transactions")
	disbursements := make([]*disbursementModel.DisbursementWithTransaction, 0)

	query := `SELECT 
		d.uuid, d.reference_id, d.merchant_id, d.bulk_id, d.amount, at.status AS transaction_status, d.beneficiary_bank_code, d.beneficiary_account_no
	FROM account_transactions at
	JOIN disbursements d ON at.reference_id = d.uuid
	WHERE 
		at.type = 'DISBURSEMENT' AND at.status = 'PENDING' 
		AND (at.updated_at BETWEEN ?  AND ?) 
		AND (at.processor_reference_id IS NOT NULL AND at.processor_reference_id != '')
		AND IFNULL(at.reason_type, '') NOT IN ('CUT_OFF_TIME', 'WAITING_MANUAL_ACTION')
	ORDER BY at.updated_at ASC`

	err := r.db.SelectContext(ctx, &disbursements, query, from, to)
	return disbursements, err
}
