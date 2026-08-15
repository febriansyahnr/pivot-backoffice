package vendor

import (
	"context"

	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *VendorRepository) Update(ctx context.Context, vendor *vendorModel.Vendor) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE vendors SET
			name = :name,
			beneficial_owner = :beneficial_owner,
			business_category = :business_category,
			avg_monthly_tpv_amount = :avg_monthly_tpv_amount,
			bank_name = :bank_name,
			bank_code = :bank_code,
			account_number = :account_number,
			account_name = :account_name,
			documents = :documents,
			status = :status,
			updated_at = :updated_at
		WHERE uuid = :uuid AND deleted_at IS NULL`

	_, err := r.db.NamedExecContext(ctx, query, vendor)
	if err != nil {
		r.logger.Error(ctx, "error when updating vendor", logger.Error(err))
		return err
	}

	return nil
}
