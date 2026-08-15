package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// FindMerchantByID is a function to find merchant by id
func (r *MerchantRepository) FindMerchantByID(ctx context.Context, id string) (*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/FindMerchantByID")
	defer segment.End()

	var data merchant.Merchant

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	query := `
		SELECT
			m.uuid, m.external_id, m.name, m.short_name, m.description, m.website, m.address, m.district_id, m.postcode, m.logo, m.merchant_email, m.merchant_phone, m.pic_email, m.pic_phone, m.kyc_status,
			m.mid, m.callback_api_key, m.callback_api_key_version, IFNULL(m.jit_api_key, '') AS jit_api_key, m.jit_api_key_version, m.parent_id, m.created_at, m.updated_at, m.deleted_at,
			m.business_type, m.business_structure, m.business_country, m.pic_name, m.pic_job_title, m.status, m.reason_status,
			m.parent_industry, m.child_industry, m.mcc, m.country_of_entity, m.digital_status,
			JSON_OBJECT('province', p.name, 'city', c.name, 'district', d.name) AS address_detail, IFNULL(u.status, 'NOT_INVITED') AS pic_invitation,
			m.third_party_screening_data, m.metadata, m.risk_level
		FROM
			merchants as m
		LEFT JOIN districts d ON d.id = m.district_id
		LEFT JOIN cities c ON c.id = d.city_id
		LEFT JOIN provinces p ON p.id = c.province_id
		LEFT JOIN users u ON u.email = m.pic_email AND u.merchant_id = m.uuid
		WHERE m.uuid = ?`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant", logger.Error(err))
		return &data, err
	}

	return &data, nil
}

func (r *MerchantRepository) FindMerchantForQrRegistration(ctx context.Context, merchantId, acquirer string) (resp *merchant.QrisMerchant, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/FindMerchantForQrRegistration")
	defer segment.End()

	resp = &merchant.QrisMerchant{
		Documents:        []merchant.QrisDocument{},
		BoardOfDirectors: []merchant.QrisBOD{},
	}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "multiple_qr_registration_table")

	rawQuery := `SELECT 
		IFNULL(qr.id, '') AS registration_id, m.uuid, m.external_id, IFNULL(m.parent_id, '') AS parent_id, m.name, m.short_name, 
		JSON_OBJECT('province', p.id, 'city', c.id, 'district', d.id, 'postcode', m.postcode, 'detail', m.address) AS address, 
		IFNULL(mp.name, '') AS parent_name, IFNULL(qr.status, '') AS qr_status, IFNULL(qrp.acquirer_merchant_id, '') AS qr_acquirer_merchant_id, m.mid,
		IFNULL(m.mcc, '') AS mcc
	FROM merchants m
	LEFT JOIN districts d ON d.id = m.district_id 
	LEFT JOIN cities c ON c.id = d.city_id 
	LEFT JOIN provinces p ON p.id = c.province_id 
	LEFT JOIN qr_registrations qr ON qr.external_id = m.external_id AND qr.acquirer = ?
	LEFT JOIN merchants mp ON mp.uuid = m.parent_id
	LEFT JOIN qr_registrations qrp ON qrp.external_id = mp.external_id AND qrp.acquirer = ? AND qrp.status = 'SUCCESS'
	WHERE
		m.uuid = ? AND m.deleted_at IS NULL;`

	if err = r.db.GetContext(ctx, resp, rawQuery, acquirer, acquirer, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(resp.AddressRaw, &resp.Address); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}

	rawQueryDoc := `SELECT type, identifier as number, location FROM merchant_documents WHERE merchant_id = ? AND deleted_at IS NULL AND status = ?;`

	err = r.db.SelectContext(ctx, &resp.Documents, rawQueryDoc, merchantId, constant.StatusApproved)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("merchant document: %v", err)
	}
	for i := range resp.Documents {
		_ = json.Unmarshal(resp.Documents[i].LocationRaw, &resp.Documents[i].Location)
	}

	rawQueryBOD := `SELECT position, identity_number, identity_file FROM merchant_bods WHERE merchant_id = ? AND deleted_at IS NULL AND status = ?;`

	err = r.db.SelectContext(ctx, &resp.BoardOfDirectors, rawQueryBOD, merchantId, constant.StatusApproved)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("merchant bod: %v", err)
	}
	for i, d := range resp.BoardOfDirectors {
		if d.Position == constant.PositionDirector {
			resp.BODCount++
		} else {
			resp.BOCCount++
		}
		_ = json.Unmarshal(
			d.IdentityFileRaw, &resp.BoardOfDirectors[i].IdentityFile,
		)
	}
	return
}

func (r *MerchantRepository) GetListMerchantFeeThatUseTiers(ctx context.Context) (result map[string][]merchant.MerchantFeeThatUseTier, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetListMerchantFeeThatUseTiers")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	merchantFees := []merchant.MerchantFeeThatUseTier{}

	rawQuery := `SELECT 
			uuid, merchant_id, reference, payment_method, tiering_type, tiering_configs
		FROM merchant_fees
		WHERE
			tiering_type IS NOT NULL
			AND (tiering_model IS NULL OR tiering_model = 'MONTHLY_ASSESSED')
			AND deleted_at IS NULL
		ORDER BY merchant_id, reference, payment_method;`
	if err = r.db.SelectContext(ctx, &merchantFees, rawQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result = map[string][]merchant.MerchantFeeThatUseTier{}

	for _, fee := range merchantFees {
		_ = fee.RawTieringConfigs.Unmarshal(&fee.TieringConfigs)

		key := fee.MerchantId
		if fee.Reference == constant.ReferencePlatformActivity {
			key += "_" + constant.ReferencePlatformActivity
		}
		result[key] = append(result[key], fee)
	}
	return
}

func (r *MerchantRepository) GetSubMerchantIdListByParentId(ctx context.Context, parentId string) (result []string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetSubMerchantIdListByParentId")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT uuid FROM merchants WHERE parent_id = ? AND deleted_at IS NULL ORDER BY created_at;`
	if err = r.db.SelectContext(ctx, &result, rawQuery, parentId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

func (r *MerchantRepository) GetAllActiveMerchantIDs(ctx context.Context) (result []string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetAllActiveMerchantIDs")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT uuid FROM merchants WHERE status = ? AND deleted_at IS NULL ORDER BY created_at;`
	if err = r.db.SelectContext(ctx, &result, rawQuery, constant.MerchantStatusActive); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

func (r *MerchantRepository) ListUnencryptedMerchantSecretsForMigration(ctx context.Context) ([]merchant.UnencryptedMerchantSecretsForMigration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/ListUnencryptedMerchantSecretsForMigration")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants,merchant_auths")

	rawQuery := `SELECT 
		m.uuid, m.name, IFNULL(m.callback_api_key, '') AS callback_api_key, m.callback_api_key_version, IFNULL(m.jit_api_key,'') AS jit_api_key, m.jit_api_key_version, ma.secret, ma.secret_version
	FROM merchants m
	JOIN merchant_auths ma ON ma.merchant_id = m.uuid 
	WHERE
		m.callback_api_key_version = 0;`

	result := []merchant.UnencryptedMerchantSecretsForMigration{}
	if err := r.db.SelectContext(ctx, &result, rawQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "failed while retrieving unencrypted merchant secret list", logger.Error(err))
		return nil, err
	}
	return result, nil
}
