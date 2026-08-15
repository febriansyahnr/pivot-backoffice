package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *MerchantRepository) DeterminePaymentFeeByMerchantIdMethodAndChannel(ctx context.Context, request *feeModel.GetFeeRequest) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/DeterminePaymentFeeByMerchantIdMethodAndChannel")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT 
		uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
	FROM (
		SELECT
			0 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel = ? AND settlement_model = ?
		UNION ALL
		SELECT 
			1 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel IS NULL AND settlement_model = ?
		UNION ALL
		SELECT
			2 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel = ? AND settlement_model IS NULL
		UNION ALL
		SELECT 
			3 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel IS NULL AND settlement_model IS NULL
	) foo
	ORDER BY priority LIMIT 1;`

	args := []any{
		request.MerchantID, constant.ReferencePayment, request.PaymentMethod, request.Channel, request.SettlementModel,
		request.MerchantID, constant.ReferencePayment, request.PaymentMethod, request.SettlementModel,
		request.MerchantID, constant.ReferencePayment, request.PaymentMethod, request.Channel,
		request.MerchantID, constant.ReferencePayment, request.PaymentMethod,
	}

	result := &merchant.MerchantFee{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if result.TieringConfigs.Valid {
		_ = result.TieringConfigs.JSONText.Unmarshal(&result.TieringConfigsObj)
	}
	return result, nil
}

func (r *MerchantRepository) DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel(ctx context.Context, merchantId, method, channel, settlementMethod string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT 
		uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
	FROM (
		SELECT
			0 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND settlement_method = ? AND channel = ? 
		UNION ALL
		SELECT
			1 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND settlement_method = ? AND channel IS NULL
	) foo
	ORDER BY priority LIMIT 1;`

	args := []any{
		merchantId, constant.ReferencePaymentFundedPayout, method, settlementMethod, channel,
		merchantId, constant.ReferencePaymentFundedPayout, method, settlementMethod,
	}

	result := &merchant.MerchantFee{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if result.TieringConfigs.Valid {
		_ = result.TieringConfigs.JSONText.Unmarshal(&result.TieringConfigsObj)
	}
	return result, nil
}

func (r *MerchantRepository) DeterminePayoutFeeByMerchantIdAndChannel(ctx context.Context, merchantId, channel, reference string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/DeterminePayoutFeeByMerchantIdAndChannel")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT
		uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
	FROM (
		SELECT
			0 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND channel = ?
		UNION ALL
		SELECT
			1 AS priority, uuid, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs, tiering_model, tiering_type, tiering_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND channel IS NULL
	) foo
	ORDER BY priority LIMIT 1;`

	args := []any{
		merchantId, reference, channel, merchantId, reference,
	}

	result := &merchant.MerchantFee{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if result.TieringConfigs.Valid {
		_ = result.TieringConfigs.JSONText.Unmarshal(&result.TieringConfigsObj)
	}
	return result, nil
}

func (r *MerchantRepository) DetermineRefundFeeByMerchantIdAndReferenceType(ctx context.Context, merchantId, referenceType string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/DetermineRefundFeeByMerchantIdAndReferenceType")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT
			amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND reference_type = ? LIMIT 1`

	args := []any{
		merchantId, constant.TypeRefund, referenceType,
	}

	result := &merchant.MerchantFee{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *MerchantRepository) DetermineTopupFeeByMerchantIdMethodAndChannel(ctx context.Context, merchantId, method, channel string) (*merchant.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/DetermineTopupFeeByMerchantIdMethodAndChannel")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `SELECT
		amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs
	FROM (
		SELECT
			0 AS priority, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel = ?
		UNION ALL
		SELECT
			1 AS priority, amount_type, amount, percentage, tax_type, tax_percentage, max_fee_amount, deduction_type, deduction_day, deduction_last_date, settlement_configs
		FROM
			merchant_fees
		WHERE merchant_id = ? AND reference = ? AND payment_method = ? AND channel IS NULL
	) foo
	ORDER BY priority LIMIT 1;`

	args := []any{
		merchantId, constant.ReferenceTopUp, method, channel,
		merchantId, constant.ReferenceTopUp, method,
	}

	result := &merchant.MerchantFee{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}
