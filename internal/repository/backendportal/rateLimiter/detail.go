package rateLimiter

import (
	"context"
	"database/sql"
	"errors"

	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RateLimiterRepository) Detail(ctx context.Context, uuid string) (*ratelimiter.RateLimitConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/rateLimiter/Detail")
	defer segment.End()

	var configuration ratelimiter.RateLimitConfiguration
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)
	query := `
		SELECT uuid, merchant_id, ` + "`limit`" + `, burst, ` + "`order`" + `, time, variable, variable_value, 
		variable_match_type, http_method, status, description, created_at, updated_at 
		FROM ` + MerchantRateLimitTable + `
		WHERE 
			uuid = ?
	`

	err := r.db.GetContext(ctx, &configuration, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding ratelimitter", logger.Error(err), logger.Any("uuid", uuid))
		return nil, err
	}
	return &configuration, err
}
