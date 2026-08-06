package ratelimiter

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *rateLimiterService) resetFailedAttempts(ctx context.Context, key, lockKey string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/resetFailedAttempts")
	defer segment.End()

	_, err := s.redis.Del(ctx, key).Result()
	if err != nil {
		s.logger.Error(ctx, "failed reset failed attempts", logger.Error(err), logger.Any("key", lockKey))
		return constant.ErrRateLimiterFailedResetAttempts
	}

	return nil
}
