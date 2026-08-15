package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *DisbursementRepository) GetDetailForCardFundedPayoutByID(ctx context.Context, id string) (*model.Disbursement, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetDetailForCardFundedPayoutByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
		uuid, reference_id, merchant_id, type, sender_name, beneficiary_bank_code,
		beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, 
		currency, amount, fee, total_amount, status, reason_type, reason_description, remark,
		metadata, created_from, created_by, approved_by, approved_at, created_at, updated_at
	FROM disbursements
	WHERE
		uuid = ? AND type = '` + constant.DisbursementTypeCardFundedPayout + `' AND deleted_at IS NULL;`

	result := model.Disbursement{}
	if err := r.db.GetContext(ctx, &result, rawQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if result.Metadata.Valid {
		_ = result.Metadata.Unmarshal(&result.MetadataObj)
	}
	return &result, nil
}
