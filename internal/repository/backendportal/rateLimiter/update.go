package rateLimiter

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *RateLimiterRepository) Update(ctx context.Context, configuration *ratelimiter.RateLimitConfiguration) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/rateLimiter/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)
	query := `
		UPDATE ` + MerchantRateLimitTable + `
		SET 
			merchant_id = :merchant_id, 
			` + "`limit`" + ` = :limit, 
			` + "`order`" + ` = :order, 
			burst = :burst,
			time = :time, 
			variable = :variable, 
			variable_value = :variable_value, 
			variable_match_type = :variable_match_type, 
			http_method = :http_method, 
			status = :status, 
			description = :description, 
			updated_at = :updated_at
		WHERE 
			uuid = :uuid
	`

	affected, err := r.db.NamedExecContext(ctx, query, configuration)
	if err != nil {
		r.logger.Error(ctx, "error when update rate limiter configuration", logger.Error(err), logger.Any("configuration", configuration))
		return err
	}

	if !affected {
		r.logger.Info(ctx, "failed when update rate limiter configuration", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}
	return nil
}
