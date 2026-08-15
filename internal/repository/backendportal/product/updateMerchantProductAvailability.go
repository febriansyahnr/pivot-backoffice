package productRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *ProductRepository) UpdateMerchantProductAvailability(ctx context.Context, req *product.UpdateMerchantSelectedProductAvailabilityRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/product/UpdateMerchantProductAvailability")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantSelectedProductTableName)
	query := `
		UPDATE ` + merchantSelectedProductTableName + ` 
		SET active = ? WHERE merchant_id = ? AND product_id = ?
		`

	_, err := r.db.ExecContext(ctx, query, req.Active, req.MerchantID, req.ProductID)
	if err != nil {
		r.logger.Error(ctx, "error when updating merchant product availability", logger.Error(err), logger.Any("request", req))
		return err
	}

	return nil
}
