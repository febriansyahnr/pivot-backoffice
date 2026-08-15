package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
)

// GetTransactionByProcessorID implements repository.IAccountTransactionRepository.
func (r *AccountTransactionRepository) GetTransactionByProcessorID(ctx context.Context, trxType, processor, processorID string) (*reconciliation.ReconTransactionModel, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetTransactionForRecon")
	defer segment.End()

	var (
		err         error
		transaction reconciliation.ReconTransactionModel
	)
	query := `
		select 
				t.uuid,
				t.type,
				COALESCE(p.amount, 0) as amount,
				t.reference_id,
				m.name as merchant_name,
				t.processor_reference,
				t.processor_reference_id,
				t.reference,
				t.channel,
				t.status,
				t.reason_type,
				t.reason_description,
				t.additional_info,
				t.transaction_timestamp
		from account_transactions t
		left join payments p on p.uuid = t.reference_id AND t.type = 'PAYMENT'
		left join merchants m on m.uuid = t.merchant_id
		where 
			t.processor_reference = ?
			and t.processor_reference_id = ?
			and t.type = ?
		limit 1
	`
	err = r.db.GetContext(
		ctx,
		&transaction,
		query,
		processor,
		processorID,
		trxType,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find account transaction for recon", logger.Error(err))
		return nil, err
	}

	return &transaction, nil
}
