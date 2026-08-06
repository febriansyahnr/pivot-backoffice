package dailyAccountTransactionRepository

import (
	"context"
	"time"

	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

// Upsert inserts or updates a new daily account transaction into the database
func (r *DailyAccountTransactionRepository) Upsert(
	ctx context.Context,
	dailyAccountTransaction *dailyAccountTransactionModel.DailyAccountTransaction,
) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/dailyAccountTransaction/Insert")
	defer span.End()

	// Set table name in context if needed by mySqlExt
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	// Set CreatedAt and UpdatedAt to current UTC time
	dailyAccountTransaction.CreatedAt = time.Now().UTC()
	dailyAccountTransaction.UpdatedAt = time.Now().UTC()

	// Define the INSERT query with named parameters
	query := `
        INSERT INTO daily_account_transactions (
            id, account_id, date, timezone, beg_balance, debit_trx, debit_amount, 
            credit_trx, credit_amount, eod_balance, created_at, updated_at
        ) VALUES (
            :id, :account_id, :date, :timezone, :beg_balance, :debit_trx, :debit_amount, 
            :credit_trx, :credit_amount, :eod_balance, :created_at, :updated_at
        )
        ON DUPLICATE KEY UPDATE
		    beg_balance = VALUES(beg_balance),
		    debit_trx = VALUES(debit_trx),
		    debit_amount = VALUES(debit_amount),
		    credit_trx = VALUES(credit_trx),
		    credit_amount = VALUES(credit_amount),
		    eod_balance = VALUES(eod_balance),
		    updated_at = VALUES(updated_at);
	`

	_, err := r.db.NamedExecContext(ctx, query, dailyAccountTransaction)
	if err != nil {
		r.logger.Error(ctx, "error when upsert daily account transaction", pdkLogger.Error(err))
		return err
	}

	return nil
}
