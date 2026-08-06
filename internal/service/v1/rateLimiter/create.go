package ratelimiter

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *rateLimiterService) Create(ctx context.Context, req *ratelimiter.CreateRateLimitConfiguration) (*ratelimiter.RateLimitConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/rateLimiter/Create")
	defer segment.End()

	configuration, err := ratelimiter.New(req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	err = s.rateLimiterRepo.Create(ctx, configuration)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrCreateRateLimiterConfiguration)
	}

	_, err = s.CacheMerchantRateLimitConfig(ctx, req.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to store rate limit cache", pdkLog.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrCreateRateLimiterConfiguration)
	}

	return configuration, nil
}
