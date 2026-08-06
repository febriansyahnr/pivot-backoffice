package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *rateLimiterService) updateFailedAttempts(ctx context.Context, key, lockKey string, timestamp time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/updateFailedAttempts")
	defer segment.End()

	addVal := int64(0)
	_, err := s.redis.Client().ZAdd(ctx, key, redis.Z{
		Member: timestamp.Unix(),
		Score:  float64(timestamp.Unix()),
	}).Result()
	if err != nil {
		addVal++
		s.logger.Error(ctx, "error when add current timestamp to redis sorted set", pdkLog.Error(err), pdkLog.Any("key", key))
	}

	startTimestamp := fmt.Sprintf("%d", timestamp.Add(time.Duration(-FailedAttemptTimeFrameInMinute)*time.Minute).Unix())
	endTimestamp := fmt.Sprintf("%d", timestamp.Unix())

	_, err = s.redis.Client().ZRemRangeByScore(ctx, key, "0", startTimestamp).Result()
	if err != nil {
		s.logger.Error(ctx, "error remove older failed attempts", pdkLog.Error(err), pdkLog.Any("startTimestamp", startTimestamp), pdkLog.Any("endTimeFrame", timestamp), pdkLog.Any("key", key))
	}

	count, err := s.redis.Client().ZCount(ctx, key, startTimestamp, endTimestamp).Result()
	if err != nil {
		s.logger.Error(ctx, "error when count existing failed attempts", pdkLog.Error(err), pdkLog.Any("startTimestamp", startTimestamp), pdkLog.Any("endTimeFrame", timestamp), pdkLog.Any("key", key))
		return constant.ErrRateLimiterFailedUpdateAttempts
	}
	if count+addVal >= MaxFailedAttempt {
		ok, err := s.redis.Client().Expire(ctx, key, LockDuration).Result()
		if err != nil {
			s.logger.Error(ctx, "error set expiry time", pdkLog.Error(err), pdkLog.Any("key", key))
			return constant.ErrRateLimiterFailedUpdateAttempts
		}

		if !ok {
			s.logger.Error(ctx, "failed set expiry time", pdkLog.Error(err), pdkLog.Any("key", key))
			return constant.ErrRateLimiterFailedUpdateAttempts
		}
	}
	return nil
}

func (s *rateLimiterService) Update(ctx context.Context, req *ratelimiter.UpdateRateLimitConfiguration) (*ratelimiter.RateLimitConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/Update")
	defer segment.End()

	config, err := s.rateLimiterRepo.Detail(ctx, req.ID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetRateLimiterConfigurationDetail)
	}
	if config == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrRateLimiterConfigurationNotFound)
	}

	err = config.Update(req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	err = s.rateLimiterRepo.Update(ctx, config)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateRateLimiterConfiguration)
	}

	_, err = s.CacheMerchantRateLimitConfig(ctx, req.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to store rate limit cache", pdkLog.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateRateLimiterConfiguration)
	}

	return config, nil
}
