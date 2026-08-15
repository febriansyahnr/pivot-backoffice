package accounttransaction_repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) FindByReference(
	ctx context.Context,
	referenceID, referenceType string,
) (*orchestratorModel.AccountTransactionWithUseCase, error) {

	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/FindByReference")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var transaction orchestratorModel.AccountTransactionWithUseCase
	query := `SELECT 
			t.uuid, 
			t.merchant_id, 
			t.reference_id, 
			t.account_id,       
			t.currency, 
			t.credit, 
			t.debit, 
			CASE 
				WHEN t.type = 'DISBURSEMENT' AND d.bulk_id IS NOT NULL THEN 'BULK_DISBURSEMENT'  
				ELSE t.type
			END AS type, 
			t.status, 
			t.reason_type, 
			t.reason_description, 
			t.remarks, 
			t.processor_reference,
			t.transaction_timestamp,
			t.settlement_at,
			t.settlement_status,
			t.settlement_model,
			t.created_at, 
			t.updated_at, 
			COALESCE(d.sender_name, '-') as sender_name,
			CASE 
				WHEN d.fee IS NOT NULL THEN d.fee 
				WHEN p.fee IS NOT NULL THEN p.fee 
				ELSE 0 
			END as fee,
			COALESCE(d.beneficiary_account_no, '-') as beneficiary_account_no,
			COALESCE(d.beneficiary_account_name, '-') as beneficiary_account_name,
			COALESCE(d.beneficiary_bank_name, '-') as beneficiary_bank_name,
			CASE 
				WHEN d.reference_id IS NOT NULL THEN d.reference_id 
				WHEN p.reference_id IS NOT NULL THEN p.reference_id 
				ELSE '-' 
			END as client_reference_id,
			d.approved_at,
			COALESCE(t.processor_reference_id, '-') as processor_reference_id,
			COALESCE(t.processor_transaction_id, '') as processor_transaction_id,
			COALESCE(c.name, 'System') as created_by,
			COALESCE(a.name, '-') as approved_by, t.additional_info
        FROM account_transactions t
        LEFT JOIN disbursements d ON d.uuid = t.reference_id AND t.type = 'DISBURSEMENT'
        LEFT JOIN payments p ON p.uuid = t.reference_id AND t.type = 'PAYMENT'
        LEFT JOIN users c ON c.uuid = d.created_by
        LEFT JOIN users a ON a.uuid = d.approved_by
        WHERE t.reference_id = ? AND t.type = ?`

	if err := r.db.GetContext(ctx, &transaction, query, referenceID, referenceType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find account transaction by reference", logger.Error(err))
		return nil, err
	}
	if transaction.AdditionalInfo.Valid {
		detail := orchestratorModel.FeeTransactionMetadataObject{}
		_ = json.Unmarshal(transaction.AdditionalInfo.JSONText, &detail)
		transaction.AdditionalInfoObj = detail
	}
	return &transaction, nil
}
