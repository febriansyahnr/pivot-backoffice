package withdrawalRepository

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *withdrawalRepository) Create(ctx context.Context, data *withdrawal.Withdrawal) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/UpdateMetadataById")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	data.RawMetadata = types.NullJSONText{Valid: true}
	data.RawMetadata.JSONText, _ = json.Marshal(data.Metadata)

	rawQuery := `INSERT INTO withdrawals (
				id, merchant_id, reference_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no,
				beneficiary_account_name, currency, amount, created_by, created_at, updated_at, type, description, metadata
		) VALUES(
			:id, :merchant_id, :reference_id, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no, 
			:beneficiary_account_name, :currency, :amount, :created_by, :created_at, :updated_at, :type, :description, :metadata
		);`
	_, err := r.db.NamedExecContext(ctx, rawQuery, data)
	return err
}
