package ratelimiter

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *rateLimiterService) validateAttempt(ctx context.Context, key, lockKey string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/validateAttempt")
	defer segment.End()

	duration, err := s.redis.Client().TTL(ctx, key).Result()
	if err != nil {
		s.logger.Error(ctx, "error get key TTL", logger.Error(err), logger.Any("key", key))
		return constant.ErrRateLimiterFailedValidate
	}
	if duration < 0 {
		return nil
	}

	return constant.ErrRateLimiterExceedFailedAttempts
}
