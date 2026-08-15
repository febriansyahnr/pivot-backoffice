package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) GetListOfMerchantsWhoHaveSubMerchant(ctx context.Context) (result []merchant.MerchantWithSubMerchantList, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetListOfMerchantsWhoHaveSubMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT 
			parent_id AS id, 
			IFNULL(JSON_ARRAYAGG(uuid), '[]') AS sub_merchants,
			(SELECT created_at FROM merchants WHERE uuid = sub.parent_id) AS created_at
		FROM merchants sub
		WHERE parent_id IS NOT NULL AND deleted_at IS NULL GROUP BY parent_id;`
	if err = r.db.SelectContext(ctx, &result, rawQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	for i := range result {
		result[i].RawSubMerchants.Unmarshal(&result[i].SubMerchants)
	}
	return
}

func (r *MerchantRepository) GetMerchantFeeListForBalanceDeduction(ctx context.Context) (result []merchant.MerchantFeeForBalanceDeduction, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantFeeListForBalanceDeduction")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT
			merchant_id, reference, IFNULL(payment_method, '') AS method, deduction_day, deduction_last_date, created_at
		FROM merchant_fees
		WHERE reference != 'PLATFORM_ACTIVITY' AND deduction_type = 'AUTOMATED' AND deleted_at IS NULL;`

	if err = r.db.SelectContext(ctx, &result, rawQuery); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
