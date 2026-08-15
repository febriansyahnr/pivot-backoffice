package productRepository

import (
	"context"
	"database/sql"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *ProductRepository) GetProductById(ctx context.Context, productId string) (*product.Product, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetProductList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, productTableName)

	var product product.Product
	query := `
		SELECT uuid, name, active, created_at, updated_at
		FROM ` + productTableName + `
		WHERE uuid=?
		LIMIT 1
		`
	err := r.db.GetContext(ctx, &product, query, productId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn(ctx, "product not found", logger.Any("productId", productId))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding product", logger.Error(err), logger.Any("productId", productId))
		return nil, err
	}

	return &product, nil
}
