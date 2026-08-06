package reportingRepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *repository) ListAccountTransactionsForMigration(ctx context.Context, startDate, endDate time.Time) ([]cdcModel.AccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/ListAccountTransactionsForMigration")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "account_transactions")

	rawQuery := `SELECT
		at.uuid, at.reference_id, merchant_id, account_id, merchant_reference_id, at.currency, credit, debit, at.type, channel, status, reason_type, reason_description, remarks, 
		at.reference, at.additional_info, settlement_at, settlement_status, settlement_model, at.created_at, at.updated_at
	FROM account_transactions at
	JOIN accounts acc ON acc.uuid = at.account_id AND acc.user_type != 'CUSTOMER'
	WHERE
		(at.created_at BETWEEN DATE_SUB(?, INTERVAL 30 DAY) AND ?) AND 
		(at.updated_at BETWEEN ? AND ?) AND status = 'SUCCESS'
	ORDER BY at.created_at;`

	transactions := []cdcModel.AccountTransaction{}
	if err := r.db.SelectContext(ctx, &transactions, rawQuery, startDate, endDate, startDate, endDate); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return transactions, nil
}
