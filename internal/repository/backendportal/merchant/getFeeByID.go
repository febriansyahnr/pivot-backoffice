package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) GetMerchantFeeByID(ctx context.Context, id string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantFeeByID")
	defer segment.End()

	var data merchant.MerchantFee

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	query := `
		SELECT
			m.uuid, m.merchant_id, m.payment_method, m.amount, m.amount_type, m.percentage, m.reference, m.deduction_type, m.tax_type, 
			m.tax_percentage, m.created_at, m.updated_at, m.deleted_at, m.max_fee_amount, m.deduction_day, m.deduction_last_date, m.settlement_configs, m.settlement_model, channel
		FROM
			merchant_fees as m
		WHERE m.uuid = ?`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant fee by id", logger.Error(err))
		return &data, err
	}

	data.SettlementConfigs.JSONText.Unmarshal(&data.SettlementConfigsObj)

	return &data, nil

}
