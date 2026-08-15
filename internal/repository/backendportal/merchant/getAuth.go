package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) GetMerchantAuthByMerchantId(ctx context.Context, merchantID string) (*merchant.MerchantAuth, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetMerchantAuthByMerchantId")
	defer segment.End()

	var data merchant.MerchantAuth

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantAuthTable)

	query := `SELECT
			m.uuid, m.merchant_id, m.secret, m.secret_version, m.created_at, m.updated_at
		FROM
			merchant_auths AS m
		WHERE m.merchant_id = ? AND m.deleted_at IS NULL;`

	if err := r.db.GetContext(ctx, &data, query, merchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn(ctx, "merchant ID not found in merchant credentials table", logger.String("merchantId", merchantID))
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding merchant auth", logger.Error(err))
		return nil, err
	}

	return &data, nil
}

func (r *MerchantRepository) GetMerchantSnapPKCS8KeyByMerchantID(ctx context.Context, merchantID string) (*merchant.MerchantAuth, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetMerchantSnapPKCS8Key")
	defer segment.End()

	var data merchant.MerchantAuth

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchant_auths")

	query := `
		SELECT
			m.uuid, m.merchant_id, m.secret, m.secret_version, m.merchant_public_key, m.snap_private_key, m.created_at, m.updated_at, m.deleted_at
		FROM
			merchant_auths as m
		WHERE m.merchant_id = ?`

	if err := r.db.GetContext(ctx, &data, query, merchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "merchant auth not found", logger.Error(err))
			return nil, err
		}

		r.logger.Error(ctx, "error when finding merchant auth", logger.Error(err))
		return &data, err
	}

	return &data, nil

}
