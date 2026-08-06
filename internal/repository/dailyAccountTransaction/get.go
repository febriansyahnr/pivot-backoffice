package dailyAccountTransactionRepository

import (
	"context"
	"database/sql"
	"errors"

	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DailyAccountTransactionRepository) FindLatestByAccountIDAndTimezone(
	ctx context.Context,
	accountID, timezone string,
) (*dailyAccountTransactionModel.DailyAccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/dailyAccountTransactionRepository/FindLatestByAccountIDAndTimezone")
	defer segment.End()

	// Set the table name for context (useful for logs/tracing)
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var data dailyAccountTransactionModel.DailyAccountTransaction

	query := `
        SELECT 
            id, account_id, date, timezone, beg_balance, debit_trx, 
            debit_amount, credit_trx, credit_amount, eod_balance, created_at
        FROM daily_account_transactions
        WHERE account_id = ? AND timezone = ?
        ORDER BY date DESC
        LIMIT 1`

	// Execute the query and fetch the result
	if err := r.db.GetContext(ctx, &data, query, accountID, timezone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn(ctx, "daily account transaction not found", pdkLogger.String("account_id", accountID), pdkLogger.String("timezone", timezone))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding latest daily account transaction", pdkLogger.String("account_id", accountID), pdkLogger.String("timezone", timezone), pdkLogger.Error(err))
		return nil, err
	}

	return &data, nil
}
