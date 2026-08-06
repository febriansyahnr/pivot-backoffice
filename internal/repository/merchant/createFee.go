package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) CreateMerchantFee(ctx context.Context, merchantFee *merchant.MerchantFee) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/CreateMerchantFee")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	query := `
			INSERT INTO
				merchant_fees (uuid, merchant_id, payment_method, amount, amount_type, percentage, reference, reference_type, deduction_type, tax_type, tax_percentage, created_at, updated_at, max_fee_amount, deduction_day, settlement_configs, settlement_model, settlement_method, channel)
			VALUES
				(:uuid, :merchant_id, :payment_method, :amount, :amount_type, :percentage, :reference, :reference_type, :deduction_type, :tax_type, :tax_percentage, :created_at, :updated_at, :max_fee_amount, :deduction_day, :settlement_configs, :settlement_model, :settlement_method, :channel)`

	affected, err := r.db.NamedExecContext(ctx, query, merchantFee)
	if err != nil {
		r.logger.Error(ctx, "error when inserting merchant fees", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting merchant fees", logger.Error(err))
		return err
	}
	return nil
}
