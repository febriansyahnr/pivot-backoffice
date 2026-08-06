package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) GetMerchantFeeXB(
	ctx context.Context, q *merchant.MerchantFeeXBQuery) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantFeeXB")
	defer segment.End()

	var data merchant.MerchantFee

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	query := `
		SELECT
			m.uuid, m.merchant_id, m.payment_method, m.amount, m.amount_type, m.percentage, m.reference, m.deduction_type, m.tax_type, 
			m.tax_percentage, m.created_at, m.updated_at, m.deleted_at, max_fee_amount, deduction_day, deduction_last_date, m.settlement_configs
		FROM
			merchant_fees as m
		WHERE m.merchant_id = ? AND m.reference = ? AND m.channel = ?`

	if err := r.db.GetContext(ctx, &data, query, q.MerchantID, q.Reference, q.Channel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant fee", logger.Error(err), logger.Any("query", query), logger.Any("merchantID", q.MerchantID), logger.Any("feeType", q.Reference), logger.String("channel", q.Channel))
		return &data, err
	}

	data.SettlementConfigs.JSONText.Unmarshal(&data.SettlementConfigsObj)

	return &data, nil
}
