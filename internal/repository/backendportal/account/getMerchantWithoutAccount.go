package account_repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountRepository) GetMerchantsWithoutAccount(ctx context.Context, request account_model.GetEntityWithoutAccountRequest) ([]*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetMerchantsWithoutAccount")
	defer segment.End()

	var (
		data []*merchant.Merchant
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableName)
	query := `
		SELECT 
		m.uuid, m.name, m.description, m.logo, m.merchant_email, m.merchant_phone, m.pic_email, m.pic_phone,
			m.mid, m.callback_api_key, m.parent_id, m.created_at, m.updated_at, m.deleted_at,
			m.business_type, m.business_structure, m.business_country, m.pic_name, m.pic_job_title
		FROM
			` + MerchantTableName + ` as m
		LEFT JOIN ` + TableName + ` 
			as a ON m.uuid = a.reference_id 
			AND a.name = ? 
			AND a.user_type = ?
		WHERE
			(m.parent_id = ? OR m.uuid = ?) AND m.deleted_at IS NULL AND a.uuid IS NULL
		LIMIT ? 
	`
	err := r.db.SelectContext(ctx, &data, query, request.Usecase, constant.UserTypeMerchant, request.MerchantID, request.MerchantID, request.Limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug(ctx, "merchant without account is not found", logger.Any("request", request))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant without account", logger.Error(err), logger.Any("request", request))
		return nil, err

	}

	return data, nil
}
