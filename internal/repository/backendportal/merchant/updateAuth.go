package merchant

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) UpdateMerchantAuth(ctx context.Context, merchantAuth *merchantModel.MerchantAuth) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateMerchantAuth")
	defer segment.End()

	query := `UPDATE merchant_auths
		SET
			secret = :secret, merchant_public_key = :merchant_public_key, snap_private_key = :snap_private_key, updated_at = :updated_at
		WHERE uuid = :uuid`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantAuthTable)

	affected, err := r.db.NamedExecContext(ctx, query, merchantAuth)
	if err != nil {
		r.logger.Error(ctx, "error when updating merchant_auths", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when updating merchant_auths", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
