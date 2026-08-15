package productRepository

import (
	"context"
	"database/sql"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *ProductRepository) GetMerchantSelectedProductById(ctx context.Context, merchantId, productId string) (*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetMerchantSelectedProductById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantSelectedProductTableName)

	var product product.MerchantWithProductName
	query := `
		SELECT 
			m.product_id, 
			p.name, 
			m.active, 
			m.created_at,
			m.updated_at
		FROM ` + merchantSelectedProductTableName + ` as m
		JOIN ` + productTableName + ` as p
		ON m.product_id = p.uuid
		WHERE m.merchant_id = ? AND m.product_id = ?
		LIMIT 1
		`
	err := r.db.GetContext(ctx, &product, query, merchantId, productId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info(ctx, "merchant product not found", logger.Any("merchantId", merchantId), logger.Any("productId", productId))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding product by id", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("productId", productId))
		return nil, err
	}

	return &product, nil
}
