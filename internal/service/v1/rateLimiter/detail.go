package ratelimiter

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *rateLimiterService) Detail(ctx context.Context, merchantId, uuid string) (*ratelimiter.RateLimitConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/rateLimiter/Detail")
	defer segment.End()

	config, err := s.rateLimiterRepo.Detail(ctx, uuid)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetRateLimiterConfigurationDetail)
	}
	if config == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrRateLimiterConfigurationNotFound)
	}
	if config.MerchantID != merchantId {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRateLimiterConfigurationID)
	}

	return config, nil
}
