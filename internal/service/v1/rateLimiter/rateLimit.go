package ratelimiter

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *rateLimiterService) RateLimitFailedAttempt(ctx context.Context, req *ratelimiter.RateLimit) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/RateLimitFailedAttempt")
	defer segment.End()

	key := fmt.Sprintf(RateLimiterKeyFormat, req.FeatureName, req.Attribute)
	lockKey := fmt.Sprintf(RateLimiterLockKeyFormat, req.FeatureName, req.Attribute)

	isAcquireLock, err := s.redis.SetNX(ctx, lockKey, true, ExclusiveLockDuration).Result()
	if err != nil {
		s.logger.Error(ctx, "error acquire redis lock during rate limit update failed attempt.", logger.Error(err))
		return constant.ErrRateLimiterFailedUpdateAttempts
	}
	if !isAcquireLock {
		// race condition
		s.logger.Debug(ctx, "failed to acquire redis lock during rate limit update failed attempt", logger.Any("key", lockKey))
		return constant.ErrRateLimiterFailedUpdateAttempts
	}
	defer func() {
		_, err := s.redis.Del(ctx, lockKey).Result()
		if err != nil {
			s.logger.Error(ctx, "failed release redis lock key during update failed attempt", logger.Error(err), logger.Any("key", lockKey))
		}
	}()

	err = s.validateAttempt(ctx, key, lockKey)
	if err != nil {
		return err
	}

	if !req.IsCheckResultCorrect {
		return s.updateFailedAttempts(ctx, key, lockKey, req.Timestamp)
	}

	return s.resetFailedAttempts(ctx, key, lockKey)
}
