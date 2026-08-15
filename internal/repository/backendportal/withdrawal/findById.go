package withdrawalRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

// FindById retrieves a withdrawal record by ID and merchant ID without joining account_transactions.
func (r *withdrawalRepository) FindById(ctx context.Context, id, merchantId string) (*withdrawal.Withdrawal, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/FindById")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	result := &withdrawal.Withdrawal{}
	rawQuery := `SELECT
			id, merchant_id, reference_id, beneficiary_bank_code, beneficiary_bank_name,
			beneficiary_account_no, beneficiary_account_name, type, description,
			currency, amount, metadata, created_by, created_at, updated_at
		FROM withdrawals
		WHERE id = ? AND merchant_id = ? AND deleted_at IS NULL`

	if err := r.db.GetContext(ctx, result, rawQuery, id, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if result.RawMetadata.Valid {
		_ = result.RawMetadata.Unmarshal(&result.Metadata)
	}

	return result, nil
}
