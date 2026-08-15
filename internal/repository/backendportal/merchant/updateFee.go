package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/jmoiron/sqlx/types"
)

func (r *MerchantRepository) UpdateMerchantFee(ctx context.Context, merchantFee *merchantModel.MerchantFee) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateMerchantFee")
	defer segment.End()

	query := `UPDATE merchant_fees
		SET
			payment_method = :payment_method, amount = :amount, amount_type = :amount_type, percentage = :percentage,
			deduction_type = :deduction_type, tax_type = :tax_type, tax_percentage = :tax_percentage, updated_at = :updated_at,
			max_fee_amount = :max_fee_amount, deduction_day = :deduction_day, settlement_configs = :settlement_configs, 
			settlement_model = :settlement_model, settlement_method = :settlement_method
		WHERE uuid = :uuid`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	affected, err := r.db.NamedExecContext(ctx, query, merchantFee)
	if err != nil {
		r.logger.Error(ctx, "error when updating merchant fees", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when updating merchant fees", logger.Error(errors.New("no rows affected")))
		return constant.ErrNoRowsAffected
	}

	return nil
}

func (r *MerchantRepository) UpdateMerchantFeeLastDeductionDate(ctx context.Context, merchantId, reference string, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateMerchantFeeLastDeductionDate")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	rawQuery := `UPDATE merchant_fees
		SET deduction_last_date = ?, updated_at = ? WHERE merchant_id = ? AND reference = ?;`

	_, err := r.db.ExecContext(ctx, rawQuery, date, time.Now().UTC(), merchantId, reference)
	return err
}

func (r *MerchantRepository) UpdateFeeTieringConfig(ctx context.Context, request *merchantModel.FeeTieringRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateFeeTieringConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchant_fees")

	tieringConfigs, _ := json.Marshal(request.Configs)

	rawQuery := `UPDATE merchant_fees 
			SET tiering_model = ?, tiering_type = ?, tiering_configs = ?, updated_at = ? WHERE uuid = ?;`
	if affected, err := r.db.ExecContext(ctx, rawQuery, request.Model, request.Type, types.JSONText(tieringConfigs), time.Now().UTC(), request.FeeId); err != nil {
		return err

	} else if !affected {
		return constant.ErrDataNotFound

	} else if request.AppliedFee == nil {
		return nil
	}
	return r.AppliedFeeFromTiers(ctx, request.FeeId, request.AppliedFee)
}

func (r *MerchantRepository) AppliedFeeFromTiers(ctx context.Context, feeId string, appliedFee *merchantModel.FeeTieringConfig) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/AppliedFeeFromTiers")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchant_fees")

	rawQuery := `UPDATE merchant_fees
			SET amount_type = :amount_type, amount = :amount, max_fee_amount = :max_fee_amount, 
				percentage = :percentage, tax_type = :tax_type, tax_percentage = :tax_percentage, updated_at = :updated_at
		WHERE uuid = :fee_id;`
	data := map[string]interface{}{
		"amount_type":    appliedFee.AmountType,
		"amount":         appliedFee.Amount,
		"max_fee_amount": appliedFee.MaxFeeAmount,
		"percentage":     appliedFee.Percentage,
		"tax_type":       appliedFee.TaxType,
		"tax_percentage": appliedFee.TaxPercentage,
		"updated_at":     time.Now().UTC(),
		"fee_id":         feeId,
	}
	_, err := r.db.NamedExecContext(ctx, rawQuery, data)
	return err
}
