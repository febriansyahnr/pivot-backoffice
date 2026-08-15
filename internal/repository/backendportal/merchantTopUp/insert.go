package merchantTopUp

import (
	"context"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantTopUp"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *merchantTopUpRepository) Create(ctx context.Context, data *model.MerchantTopUp) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchantTopUp/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `INSERT INTO merchant_top_up_references (uuid, merchant_id, account_name, payment_method_id, reference_number, created_at, updated_at)
		VALUES (:uuid, :merchant_id, :account_name, :payment_method_id, :reference_number, :created_at, :updated_at)`

	if _, err := r.db.NamedExecContext(ctx, query, data); err != nil {
		r.logger.Error(ctx, "error when inserting merchant top up reference", logger.Error(err))
		return err
	}
	return nil
}
