package productRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *ProductRepository) AddMerchantSelectedProduct(ctx context.Context, req *product.MerchantSelectedProduct) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/AddMerchantSelectedProduct")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantSelectedProductTableName)
	query := `
		INSERT INTO ` + merchantSelectedProductTableName + `
		(
			uuid,
			merchant_id,
			product_id,
			active,
			created_at,
			updated_at
		) VALUES (
			:uuid,
			:merchant_id,
			:product_id,
			:active,
			:created_at,
			:updated_at 
		)
	`
	affected, err := r.db.NamedExecContext(ctx, query, req)
	if err != nil {
		r.logger.Error(ctx, "error when insert merchant selected product", logger.Error(err), logger.Any("request", req))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when insert merchant selected product", logger.Error(err), logger.Any("request", req))
		return err
	}

	return nil
}
