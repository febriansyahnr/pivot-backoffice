package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) GetMerchantsByIDs(ctx context.Context, merchantIDs []string) ([]*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantsByIDs")
	defer segment.End()

	var data []*merchant.Merchant
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)
	merchantIdsParam := strings.Join(merchantIDs, "','")
	merchantIdsParam = fmt.Sprintf("'%s'", merchantIdsParam)

	query := `
		SELECT
			m.uuid, m.external_id, m.name, m.short_name, m.description, m.address, m.postcode, m.logo, m.merchant_email, m.merchant_phone, m.pic_email, m.pic_phone,
			m.mid, m.callback_api_key, m.parent_id, m.created_at, m.updated_at, m.deleted_at,
			m.business_type, m.business_structure, m.business_country, m.pic_name, m.pic_job_title,
			m.parent_industry, m.child_industry, m.mcc, m.country_of_entity, m.digital_status
		FROM
			` + merchantsTable + ` as m
		WHERE m.uuid IN (%s)`
	query = fmt.Sprintf(query, merchantIdsParam)

	if err := r.db.SelectContext(ctx, &data, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "merchant not found", logger.Error(err), logger.Any("merchantIds", merchantIDs))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchants", logger.Error(err), logger.Any("merchantIds", merchantIDs))
		return data, err
	}

	return data, nil

}
