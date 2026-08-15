package accounttransaction_repository

import (
	"context"
	"database/sql"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountTransactionRepository) GetLedgerDetail(ctx context.Context, referenceId string) ([]orchestratorModel.AccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetLedgerDetail")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
		SELECT 
			uuid,
			reference_id,
			merchant_id,
			account_id, 
			currency,
			credit,
			debit,
			type,
			channel,
			status, 	
			reason_type, 
			reason_description,
			remarks,
			reference,
			processor_reference,
			processor_reference_id,
			processor_transaction_id,
			transaction_timestamp,
			additional_info,
			created_at,
			updated_at,
			deleted_at,
			settlement_at,
			settlement_status,
			settlement_model
		FROM ` + tableName + `
		WHERE 
			reference_id = ?
			AND 
			deleted_at IS NULL;
	`
	var ledgerDetail []orchestratorModel.AccountTransaction
	err := r.db.SelectContext(ctx, &ledgerDetail, query, referenceId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return ledgerDetail, nil
}
