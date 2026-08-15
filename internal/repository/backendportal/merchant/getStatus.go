package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) FindStatusByID(ctx context.Context, id string) (*merchant.MerchantStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/FindStatusByID")
	defer segment.End()

	var data merchant.MerchantStatusResponse
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantFeeTable)

	query := `
		SELECT
			uuid, status, kyc_status, reason_status
		FROM
			merchants
		WHERE uuid = ?`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "merchant not found", logger.Error(err))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant status by id", logger.Error(err))
		return &data, err
	}

	return &data, nil
}
