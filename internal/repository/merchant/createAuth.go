package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) CreateMerchantAuth(ctx context.Context, merchantAuth *merchant.MerchantAuth) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/CreateMerchantAuth")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantAuthTable)

	query := `
			INSERT INTO
				merchant_auths (uuid, merchant_id, secret, merchant_public_key, snap_private_key, created_at, updated_at, secret_version)
			VALUES
				(:uuid, :merchant_id, :secret, :merchant_public_key, :snap_private_key, :created_at, :updated_at, :secret_version)`

	affected, err := r.db.NamedExecContext(ctx, query, merchantAuth)
	if err != nil {
		r.logger.Error(ctx, "error when inserting merchant auths", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting merchant auths", logger.Error(err))
		return err
	}
	return nil
}
