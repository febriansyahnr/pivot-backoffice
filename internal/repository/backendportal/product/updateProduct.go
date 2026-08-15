package productRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *ProductRepository) UpdateProductAvailability(ctx context.Context, req *product.UpdateProductRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/UpdateProductAvailability")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, productTableName)
	query := `
		UPDATE ` + productTableName + ` 
		SET active = ? WHERE uuid = ?
		`

	_, err := r.db.ExecContext(ctx, query, req.Active, req.ID)
	if err != nil {
		r.logger.Error(ctx, "error when updating product availability", logger.Error(err), logger.Any("request", req))
		return err
	}

	return nil
}
