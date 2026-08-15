package rateLimiter

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *RateLimiterRepository) List(ctx context.Context, req *ratelimiter.MerchantRateLimitRequest) ([]*ratelimiter.RateLimitConfiguration, int64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchantRateLimit/List")
	defer segment.End()

	var (
		list        = []*ratelimiter.RateLimitConfiguration{}
		whereClause []string
		whereArgs   []interface{}
		errG        = new(errgroup.Group)
		total       int64
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)

	countQuery := `SELECT COUNT(1) FROM ` + MerchantRateLimitTable
	query := `
		SELECT uuid, merchant_id, ` + "`limit`" + `, burst, ` + "`order`" + `, time, variable, variable_value, 
		variable_match_type, http_method, status, description, created_at, updated_at 
		FROM ` + MerchantRateLimitTable

	if req.MerchantID != "" {
		whereClause = append(whereClause, "merchant_id = ?")
		whereArgs = append(whereArgs, req.MerchantID)
	}
	if req.Status != "" {
		whereClause = append(whereClause, "status = ?")
		whereArgs = append(whereArgs, req.Status)
	}
	if req.Variable != "" {
		whereClause = append(whereClause, "variable = ?")
		whereArgs = append(whereArgs, req.Variable)
	}
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
		countQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	countQuery = r.db.Rebind(countQuery)
	countArgs := whereArgs
	errG.Go(func() error {
		err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
		if err != nil {
			r.logger.Error(ctx, "error counting merchant rate limits", logger.Error(err), logger.Any("query", countQuery), logger.Any("req", req))
			return err
		}
		return nil
	})

	query += " ORDER BY `order` ASC"
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query += " LIMIT ? OFFSET ?"
		whereArgs = append(whereArgs, req.PageSize, offset)
	}
	query = r.db.Rebind(query)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &list, query, whereArgs...)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			r.logger.Error(ctx, "error fetching merchant rate limits", logger.Error(err), logger.Any("query", query), logger.Any("req", req))
			return err
		}
		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetMerchantConfigs retrieves the rate limit configurations for a given merchant from the database.
// It returns a slice of MerchantRateLimitConfig pointers or an error if something goes wrong.
func (r *RateLimiterRepository) GetMerchantConfigs(ctx context.Context, merchantID string) (*[]ratelimiter.MerchantRateLimitConfig, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/rateLimiter/GetMerchantConfigs")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)

	configs := &[]ratelimiter.MerchantRateLimitConfig{}
	query := "SELECT uuid, variable, variable_value, variable_match_type, `time`, `limit`, burst, http_method FROM " + MerchantRateLimitTable + " WHERE merchant_id = ? AND deleted_at IS NULL ORDER BY `order` ASC"

	err := r.db.SelectContext(ctx, configs, query, merchantID)
	if err != nil {
		r.logger.Error(ctx, "error fetching rate limit configurations", logger.Error(err), logger.String("merchantID", merchantID))
		return nil, err
	}

	if len(*configs) == 0 {
		return nil, nil
	}

	return configs, nil
}
