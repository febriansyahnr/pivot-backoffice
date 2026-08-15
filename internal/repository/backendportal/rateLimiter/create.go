package rateLimiter

import (
	"context"

	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RateLimiterRepository) Create(ctx context.Context, configuration *ratelimiter.RateLimitConfiguration) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/rateLimiter/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)

	query := `
		INSERT INTO ` + MerchantRateLimitTable + ` (
			uuid, merchant_id, ` + "`limit`" + `, burst, ` + "`order`" + `, time, variable, variable_value, 
			variable_match_type, http_method, status, description, created_at, updated_at
		) VALUES (
			:uuid, :merchant_id, :limit, :burst, :order, :time, :variable, :variable_value, 
			:variable_match_type, :http_method, :status, :description, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, configuration)
	if err != nil {
		r.logger.Error(ctx, "error when inserting rate limit configuration", logger.Error(err), logger.Any("configuration", configuration))
		return err
	}
	return nil
}
