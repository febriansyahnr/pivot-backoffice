package merchant

import (
	"context"
	"errors"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) Create(ctx context.Context, merchant *merchantModel.Merchant) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/Create")
	defer segment.End()

	query := `
			INSERT
				INTO
					merchants (uuid, name, description, website, logo, merchant_email, merchant_phone,
							 business_type, business_structure, business_country, pic_name,
							 pic_email, pic_phone, pic_job_title, mid, status, risk_level, callback_api_key, jit_api_key, kyc_status,
							 parent_id, created_at, updated_at, deleted_at, external_id, short_name, address, district_id, postcode, transaction_configs,
							 parent_industry, child_industry, mcc, country_of_entity, digital_status, callback_api_key_version, jit_api_key_version)
				VALUES
				    (:uuid, :name, :description, :website, :logo, :merchant_email, :merchant_phone,
				     :business_type, :business_structure, :business_country, :pic_name,
					 :pic_email, :pic_phone, :pic_job_title, :mid, :status, :risk_level, :callback_api_key, :jit_api_key, :kyc_status,
					 :parent_id, :created_at, :updated_at, :deleted_at, :external_id, :short_name, :address, :district_id, :postcode, :transaction_configs,
					 :parent_industry, :child_industry, :mcc, :country_of_entity, :digital_status, :callback_api_key_version, :jit_api_key_version)`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	affected, err := r.db.NamedExecContext(ctx, query, merchant)
	if err != nil {
		r.logger.Error(ctx, "error when inserting merchant", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting merchant", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}
	return nil
}
