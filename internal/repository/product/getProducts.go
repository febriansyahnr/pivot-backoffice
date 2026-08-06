package productRepository

import (
	"context"
	"database/sql"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *ProductRepository) GetProductList(ctx context.Context) ([]*product.Product, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/GetProductList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, productTableName)

	var products []*product.Product
	query := `SELECT uuid, name, active, created_at, updated_at FROM ` + productTableName
	err := r.db.SelectContext(ctx, &products, query)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn(ctx, "product not found")
			return products, nil
		}

		r.logger.Error(ctx, "error when finding product", logger.Error(err))
		return nil, err
	}

	return products, nil
}
