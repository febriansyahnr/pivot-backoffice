package productRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *ProductRepository) GetMerchantSelectedProductByName(ctx context.Context, merchantId, productName string) (*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetMerchantSelectedProductByName")
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
		ON m.product_id = p.uuid AND p.active = true
		WHERE m.merchant_id = ? AND p.name = ? 
		LIMIT 1
		`
	err := r.db.GetContext(ctx, &product, query, merchantId, productName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding product", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("productName", productName))
		return nil, err
	}
	return &product, nil
}
