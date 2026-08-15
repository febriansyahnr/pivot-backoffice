package vendor

import (
	"context"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/vendor"
)

func (r *VendorRepository) Create(ctx context.Context, vendor *vendorModel.Vendor) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO vendors (
			uuid, merchant_id, name, beneficial_owner, business_category, avg_monthly_tpv_amount,
			bank_name, bank_code, account_number, account_name, documents, status,
			created_at, updated_at
		) VALUES (
			:uuid, :merchant_id, :name, :beneficial_owner, :business_category, :avg_monthly_tpv_amount,
			:bank_name, :bank_code, :account_number, :account_name, :documents, :status,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, vendor)
	if err != nil {
		r.logger.Error(ctx, "error when creating vendor", logger.Error(err))
		return err
	}

	return nil
}
