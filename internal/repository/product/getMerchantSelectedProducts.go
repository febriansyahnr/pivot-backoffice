package productRepository

import (
	"context"
	"database/sql"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *ProductRepository) GetMerchantSelectedProducts(ctx context.Context, merchantId string) ([]*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetMerchantSelectedProducts")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantSelectedProductTableName)

	var products []*product.MerchantWithProductName
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
		WHERE m.merchant_id = ?
		`
	err := r.db.SelectContext(ctx, &products, query, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn(ctx, "merchant products not found", logger.Any("merchantId", merchantId))
			return products, nil
		}

		r.logger.Error(ctx, "error when finding product", logger.Error(err), logger.Any("merchantId", merchantId))
		return nil, err
	}

	return products, nil
}

// GetMerchantActiveProducts retrieves a list of active products for a given merchant.
// It joins the merchant selected products table and the products table to fetch the product details.
func (r *ProductRepository) GetMerchantActiveProducts(ctx context.Context, merchantID string) ([]*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetMerchantActiveProducts")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantSelectedProductTableName)

	var products []*product.MerchantWithProductName
	query := `
		SELECT 
			m.product_id, 
			p.name, 
			m.active, 
			m.created_at,
			m.updated_at
		FROM ` + merchantSelectedProductTableName + ` as m
		JOIN ` + productTableName + ` as p
		ON m.product_id = p.uuid AND m.active = 1
		WHERE m.merchant_id = ?
		`
	err := r.db.SelectContext(ctx, &products, query, merchantID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		r.logger.Error(ctx, "error when finding product", logger.Error(err), logger.String("merchantId", merchantID))
		return nil, err
	}

	return products, nil
}
