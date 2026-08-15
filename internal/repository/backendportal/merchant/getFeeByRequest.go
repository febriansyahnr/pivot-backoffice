package merchant

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) GetMerchantFeeByRequest(ctx context.Context, request *merchant.GetMerchantFeeRequest) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantFeeByRequest")
	defer segment.End()

	var (
		data         merchant.MerchantFee
		conditionals []string
		valueParams  []interface{}
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)
	query := `
		SELECT
			m.uuid, m.merchant_id, m.payment_method, m.amount, m.amount_type, m.percentage, m.reference, m.deduction_type, m.tax_type, 
			m.tax_percentage, m.created_at, m.updated_at, m.deleted_at, m.max_fee_amount, m.deduction_day, m.deduction_last_date, m.settlement_configs,
			m.settlement_model, m.settlement_method
		FROM
			merchant_fees as m
	`

	if request.ID != "" {
		conditionals = append(conditionals, "m.uuid = ?")
		valueParams = append(valueParams, request.ID)
	}

	if request.MerchantID != "" {
		conditionals = append(conditionals, "m.merchant_id = ?")
		valueParams = append(valueParams, request.MerchantID)
	}

	if request.AmountType != "" {
		conditionals = append(conditionals, "m.amount_type = ?")
		valueParams = append(valueParams, request.AmountType)
	}

	if request.Reference != "" {
		conditionals = append(conditionals, "m.reference = ?")
		valueParams = append(valueParams, request.Reference)
	}

	if request.ReferenceType != "" {
		conditionals = append(conditionals, "m.reference_type = ?")
		valueParams = append(valueParams, request.ReferenceType)
	}

	if request.PaymentMethod != "" {
		conditionals = append(conditionals, "m.payment_method = ?")
		valueParams = append(valueParams, request.PaymentMethod)
	}

	if request.Channel != "" {
		conditionals = append(conditionals, "m.channel = ?")
		valueParams = append(valueParams, request.Channel)

	} else {
		// Because channel is an additional feature, when the function is called without including the channel, the default value assigned will be NULL.
		conditionals = append(conditionals, "m.channel IS NULL")
	}

	if request.SettlementModel != "" {
		conditionals = append(conditionals, "m.settlement_model = ?")
		valueParams = append(valueParams, request.SettlementModel)
	} else {
		// Because settlementModel is an additional feature, when the function is called without including the settlementModel, the default value assigned will be NULL.
		conditionals = append(conditionals, "m.settlement_model IS NULL")
	}

	if request.SettlementMethod != "" {
		conditionals = append(conditionals, "m.settlement_method = ?")
		valueParams = append(valueParams, request.SettlementModel)
	} else {
		conditionals = append(conditionals, "m.settlement_method IS NULL")
	}

	if len(conditionals) > 0 {
		query += "WHERE " + strings.Join(conditionals, " AND ")
	}

	query += " LIMIT 1"

	if err := r.db.GetContext(ctx, &data, query, valueParams...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant fee", logger.Error(err), logger.Any("request", request), logger.Any("query", query))
		return nil, err
	}

	data.SettlementConfigs.JSONText.Unmarshal(&data.SettlementConfigsObj)

	return &data, nil

}
