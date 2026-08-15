package accounttransaction_repository

import (
	"context"

	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) GetPastDueSettlementTransactions(ctx context.Context, request *orchestrator_model.GetPastDueSettlementTransactionsRequest) ([]*orchestrator_model.AccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetPastDueSettlementTransactions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var data []*orchestrator_model.AccountTransaction
	rawQuery := `SELECT at2.uuid, 
						at2.reference_id, 
						at2.merchant_id,
						at2.merchant_reference_id,
						at2.account_id,
						at2.currency,
						at2.credit,
						at2.debit,
						at2.reference,
						at2.type,
						at2.channel,
						at2.status,
						at2.reason_type,
						at2.reason_description,
						at2.remarks,
						at2.processor_reference,
						at2.processor_reference_id,
						at2.processor_transaction_id,
						at2.transaction_timestamp,
						at2.additional_info,
						at2.created_at,
						at2.updated_at,
						at2.deleted_at,
						at2.settlement_at,
						at2.settlement_status,
						at2.settlement_model
				FROM ` + tableName + ` as at2 
				WHERE at2.reference_id = ? 
				AND at2.status = 'SUCCESS' 
				AND at2.settlement_status = 'PENDING'
				AND 
				(
					(
						at2.additional_info ->> '$.settlementDetail.estimateSettlementAt' IS NULL AND 
						DATE_ADD(at2.updated_at, INTERVAL CAST(SUBSTRING(at2.additional_info->>'$.settlementDetail.type', 3) AS SIGNED) DAY) <= ? 
					)
					OR
					(
						at2.additional_info ->> '$.settlementDetail.estimateSettlementAt' IS NOT NULL AND 
						at2.updated_at >= at2.additional_info ->> '$.settlementDetail.estimateSettlementAt'
					)
				)
			`

	if err := r.db.SelectContext(ctx, &data, rawQuery, request.ReferenceID, request.Datetime); err != nil {
		r.logger.Error(ctx, "error when get past due settlement transactions", logger.Error(err))
		return nil, err
	}

	return data, nil
}
