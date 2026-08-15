package accounttransaction_repository

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

// Create implements repository.IAccountTransactionRepository.
func (r *AccountTransactionRepository) Create(
	ctx context.Context,
	accountTransaction *orchestrator_model.AccountTransaction,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	// Force value createdAt & updatedAt
	accountTransaction.CreatedAt = time.Now().UTC()
	accountTransaction.UpdatedAt = time.Now().UTC()

	query := `
        INSERT INTO account_transactions (
            uuid, reference_id, merchant_id, account_id, 
            currency, credit, debit, type, reference, channel,
            status, reason_type, reason_description, remarks, transaction_timestamp, settlement_at, settlement_status, settlement_model, additional_info, created_at, updated_at,
			processor_reference, processor_reference_id, processor_transaction_id, merchant_reference_id
        ) VALUES (
            :uuid, :reference_id, :merchant_id, :account_id, 
            :currency, :credit, :debit, :type, :reference, :channel,
            :status, :reason_type, :reason_description, :remarks, :transaction_timestamp, :settlement_at, :settlement_status, :settlement_model, :additional_info, :created_at, :updated_at,
			:processor_reference, :processor_reference_id, :processor_transaction_id, :merchant_reference_id
        )`

	affected, err := r.db.NamedExecContext(ctx, query, accountTransaction)
	if err != nil {
		r.logger.Error(ctx, "error when inserting account transaction", logger.Error(err))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting account transaction", logger.Error(err))
		return err
	}

	return nil
}
