package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

// ValidateMerchantRateLimit validates the rate limit for a merchant based on the provided request.
// It retrieves the rate limit configurations from Redis and enforces the rate limit based on the configurations.
func (s *rateLimiterService) ValidateMerchantRateLimit(ctx context.Context, req ratelimiter.MerchantRateLimitRequest) (*ratelimiter.MerchantRateLimitHeaderMetadata, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/ValidateMerchantRateLimit")
	defer segment.End()

	var (
		err       error
		bytes     []byte
		configs   = &[]ratelimiter.MerchantRateLimitConfig{}
		metadata  *ratelimiter.MerchantRateLimitHeaderMetadata
		hasConfig bool
	)

	err = s.redis.Get(ctx, fmt.Sprintf(MerchantRateLimitKeyConfig, req.MerchantID)).Scan(&bytes)
	if err == redisExt.ErrNil {
		configs, err = s.CacheMerchantRateLimitConfig(ctx, req.MerchantID)

		if err != nil {
			s.logger.Error(ctx, "failed to get rate limit cache configs", pdkLog.Error(err), pdkLog.String("merchantID", req.MerchantID))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("failed to get configs: %w", err))
		}
	} else if err != nil {
		s.logger.Error(ctx, "failed to get rate limit cache configs", pdkLog.Error(err), pdkLog.String("merchantID", req.MerchantID))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("failed to get configs: %w", err))
	} else if len(bytes) != 0 {
		err = json.Unmarshal(bytes, configs)
		if err != nil {
			s.logger.Error(ctx, "failed to unmarshal rate limit configs", pdkLog.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("failed to unmarshal rate limit configs: %w", err))
		}
	}

	for _, cfg := range *configs {
		if cfg.IsExactType() {
			if cfg.Variable == constant.RateLimitVariablePath && cfg.VariableValue != req.Path {
				continue
			}
		}

		if cfg.IsPrefixType() {
			if cfg.Variable == constant.RateLimitVariablePath && !strings.HasPrefix(req.Path, cfg.VariableValue) {
				continue
			}
		}

		if cfg.HTTPMethod != "" && cfg.HTTPMethod != req.HTTPMethod {
			continue
		}

		hasConfig = true
		limiterCfg := &redisExt.Limit{
			Rate:   cfg.Limit,
			Burst:  cfg.Burst,
			Period: cfg.GetDuration(),
		}

		rateLimitKey := fmt.Sprintf(MerchantRateLimitKey, req.MerchantID, cfg.UUID)
		metadata, err = s.Eval(ctx, rateLimitKey, limiterCfg)
		if err != nil {
			return metadata, err
		}
	}

	ffCtx := ffcontext.NewEvaluationContext(req.MerchantID)
	ffCtx.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, req.MerchantID)
	ffCtx.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)

	isEnabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyEnableMerchantDefaultRateLimitMiddleware, ffCtx, false)

	if !hasConfig && isEnabled {
		cfg := ratelimiter.MerchantRateLimitConfig{
			Limit: int(s.config.RateLimit.Limit),
			Time:  s.config.RateLimit.Time,
		}

		limiterCfg := &redisExt.Limit{
			Rate:   cfg.Limit,
			Burst:  cfg.Limit,
			Period: cfg.GetDuration(),
		}

		rateLimitKey := fmt.Sprintf(MerchantRateLimitDefaultKey, req.MerchantID)
		metadata, err = s.Eval(ctx, rateLimitKey, limiterCfg)
		if err != nil {
			return metadata, err
		}
		return metadata, nil
	}

	return metadata, nil
}

// CacheMerchantRateLimitConfig retrieves the rate limit configurations for a given merchant from the repository,
// stores them in the cache, and returns the configurations. If no configurations are found, it returns nil.
func (s *rateLimiterService) CacheMerchantRateLimitConfig(ctx context.Context, merchantID string) (*[]ratelimiter.MerchantRateLimitConfig, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/ratelimiter/CacheMerchantRateLimitConfig")
	defer segment.End()

	configs, err := s.rateLimiterRepo.GetMerchantConfigs(ctx, merchantID)
	if err != nil {
		return configs, err
	}

	// if no configurations are found, should generate empty array
	// it will process the request as usual because the cache should exist
	// event though the merchant does not have any rate limit configurations
	if configs == nil {
		configs = &[]ratelimiter.MerchantRateLimitConfig{}
	}

	key := fmt.Sprintf(MerchantRateLimitKeyConfig, merchantID)
	bytes, _ := json.Marshal(configs) // cannot test the failed condition due struct property

	res := s.redis.Set(ctx, key, string(bytes), 0)
	if res.Err() != nil {
		s.logger.Error(ctx, "failed to store merchant-rate-limit-config cache", pdkLog.String("merchant_id", merchantID), pdkLog.Error(res.Err()))
		return nil, res.Err()
	}

	return configs, nil
}

// Eval evaluates the rate limit for a given key and configuration.
// It returns metadata about the rate limit and an error if the rate limit is exceeded or if there is an issue setting up the limiter.
func (s *rateLimiterService) Eval(ctx context.Context, key string, cfg *redisExt.Limit) (*ratelimiter.MerchantRateLimitHeaderMetadata, error) {
	res, err := s.limiter.Allow(ctx, key, cfg)
	if err != nil {
		s.logger.Error(ctx, "failed to setup limiter", pdkLog.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	metadata := &ratelimiter.MerchantRateLimitHeaderMetadata{
		RateLimitLimit:     res.Limit.Rate,
		RateLimitRemaining: res.Remaining,
		RateLimitReset:     res.ResetAfter.Milliseconds(),
	}

	if res.Allowed <= 0 {
		return metadata, pkgErrs.New(response.HttpErrRequestLimitExceeded, errors.New("request limit exceeded"))
	}

	return metadata, nil
}
