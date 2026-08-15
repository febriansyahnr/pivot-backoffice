package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) GetMerchantFeeByMerchantIDAndType(
	ctx context.Context, merchantID, feeType string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantFeeByMerchantIDAndType")
	defer segment.End()

	var data merchant.MerchantFee

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	query := `
		SELECT
			m.uuid, m.merchant_id, m.payment_method, m.amount, m.amount_type, m.percentage, m.reference, m.deduction_type, m.tax_type, 
			m.tax_percentage, m.created_at, m.updated_at, m.deleted_at, max_fee_amount, deduction_day, deduction_last_date, m.settlement_configs,
			m.settlement_model
		FROM
			merchant_fees as m
		WHERE m.merchant_id = ? AND m.reference = ?`

	if err := r.db.GetContext(ctx, &data, query, merchantID, feeType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant fee", logger.Error(err), logger.Any("query", query), logger.Any("merchantID", merchantID), logger.Any("feeType", feeType))
		return &data, err
	}

	data.SettlementConfigs.JSONText.Unmarshal(&data.SettlementConfigsObj)

	return &data, nil
}
