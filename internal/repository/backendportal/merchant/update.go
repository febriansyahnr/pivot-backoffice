package merchant

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) Update(ctx context.Context, data *merchant.Merchant) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/Update")
	defer segment.End()

	query := `
        UPDATE merchants
        SET name = :name,
            short_name = :short_name,
            description = :description,
            website = :website,
            logo = :logo,
            merchant_email = :merchant_email,
            merchant_phone = :merchant_phone,
            business_type = :business_type,
            business_structure = :business_structure,
            business_country = :business_country,
            pic_name = :pic_name,
            pic_email = :pic_email,
            pic_phone = :pic_phone,
            pic_job_title = :pic_job_title,
            mid = :mid,
            callback_api_key = :callback_api_key,
            jit_api_key = :jit_api_key,
			jit_api_key_version = :jit_api_key_version,
            parent_id = :parent_id,
            updated_at = :updated_at,
            address = :address,
            district_id = :district_id,
            postcode = :postcode,
            status = :status,
            reason_status = :reason_status,
            parent_industry = :parent_industry,
            child_industry = :child_industry,
            mcc = :mcc,
            country_of_entity = :country_of_entity,
            digital_status = :digital_status,
            risk_level = :risk_level,
			metadata = :metadata
        WHERE uuid = :uuid`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	_, err := r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when updating merchant", logger.Error(err))
		return err
	}

	return nil
}

func (r *MerchantRepository) UpdateCallbackApiKey(ctx context.Context, merchantId, apiKey string, version uint) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateCallbackApiKey")
	defer segment.End()

	query := `UPDATE merchants
		SET
			callback_api_key = ?, callback_api_key_version = ?, updated_at = ?
		WHERE uuid = ?;`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	if _, err := r.db.ExecContext(ctx, query, apiKey, version, time.Now().UTC(), merchantId); err != nil {
		r.logger.Error(ctx, "error when updating callback api key of merchant", logger.Error(err))
		return err
	}
	return nil
}

func (r *MerchantRepository) UpdateStatusByID(ctx context.Context, status, reasonStatus, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateStatusByID")
	defer segment.End()

	query := `
		UPDATE merchants
		SET
			status = ?,
			reason_status = ?,
			updated_at = ?
		WHERE uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	_, err := r.db.ExecContext(ctx, query, status, reasonStatus, time.Now().UTC(), id)
	if err != nil {
		r.logger.Error(ctx, "error when updating status of merchant", logger.Error(err))
		return err
	}

	return nil
}

func (r *MerchantRepository) UpdateThirdPartyScreeningData(ctx context.Context, merchantID string, screeningData types.NullJSONText) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateThirdPartyScreeningData")
	defer segment.End()

	query := `
		UPDATE merchants
		SET
			third_party_screening_data = ?,
			updated_at = ?
		WHERE uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	_, err := r.db.ExecContext(ctx, query, screeningData, time.Now().UTC(), merchantID)
	if err != nil {
		r.logger.Error(ctx, "error when updating third party screening data of merchant", logger.Error(err))
		return err
	}

	return nil
}

func (r *MerchantRepository) MigrateMerchantSecretsToEncryption(ctx context.Context, merchant merchant.MigrateMerchantSecretsToEncryption) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/MigrateMerchantSecretsToEncryption")
	defer segment.End()

	rawQuery := `UPDATE
		merchants m
	JOIN
		merchant_auths ma ON ma.merchant_id = m.uuid
	SET
		m.callback_api_key = :callback_api_key, m.callback_api_key_version = :callback_api_key_version, 
		m.jit_api_key = :jit_api_key, m.jit_api_key_version = :jit_api_key_version, 
		ma.secret = :secret, ma.secret_version = :secret_version, m.updated_at = NOW(), ma.updated_at = NOW()
	WHERE
		m.uuid = :uuid;`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	_, err := r.db.NamedExecContext(ctx, rawQuery, merchant)
	return err
}
