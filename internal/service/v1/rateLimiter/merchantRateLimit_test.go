package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	appConfig "github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	rateLimitModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockRedisExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestValidateMerchantRateLimit(t *testing.T) {
	logger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	mockRedis := mockRedisExt.NewIRedisExt(t)
	mockRateLimitRepo := mockRepo.NewIRateLimiterRepository(t)
	mockLimiter := mockRedisExt.NewILimiter(t)
	ffContentConfig := `
backend-portal-merchant-rate-limit-middleware:
  variations:
    ON: true
    OFF: false
  defaultRule:
    variation: OFF
backend-portal-merchant-default-rate-limit-middleware:
  variations:
    ON: true
    OFF: false
  defaultRule:
    variation: OFF`
	f, err := os.CreateTemp(os.TempDir(), "mock-cfg-*.yaml")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(f.Name())) }()
	defer func() { require.NoError(t, f.Close()) }()

	_, err = f.WriteString(ffContentConfig)
	require.NoError(t, err)

	err = ffclient.Init(ffclient.Config{
		FileFormat: "YAML",
		Retriever: &fileretriever.Retriever{
			Path: f.Name(),
		},
	})
	require.NoError(t, err)
	defer ffclient.Close()

	traceID := "trace-id"
	path := "/api/v1/merchant"
	merchantID := "valid-merchant-id"
	configUUID := "valid-config-uuid"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)
	merchantCacheConfigKey := fmt.Sprintf(MerchantRateLimitKeyConfig, merchantID)
	merchantCacheKey := fmt.Sprintf(MerchantRateLimitKey, merchantID, configUUID)
	config := rateLimitModel.MerchantRateLimitConfig{
		UUID:              configUUID,
		VariableValue:     path,
		Limit:             10,
		Burst:             10,
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
		Time:              constant.RateLimitConfigTimeMinute,
		Variable:          constant.RateLimitVariablePath,
	}

	limiterCfg := &redisExt.Limit{
		Rate:   config.Limit,
		Burst:  config.Limit,
		Period: config.GetDuration(),
	}

	limiterResult := &redis_rate.Result{
		Limit:      *limiterCfg,
		Remaining:  9,
		Allowed:    1,
		ResetAfter: config.GetDuration() - 10,
	}

	service := New(
		logger, mockRedis, mockRateLimitRepo, WithRedisLimiter(mockLimiter), WithConfig(&appConfig.Config{
			Environment: "test",
			RateLimit: appConfig.RateLimitConfig{
				Limit: 10,
				Time:  constant.RateLimitConfigTimeMinute,
			},
		}),
	)

	optsBytes, _ := json.Marshal([]rateLimitModel.MerchantRateLimitConfig{
		config,
	})

	optsPrefixBytes, _ := json.Marshal([]rateLimitModel.MerchantRateLimitConfig{
		{
			VariableValue:     "/other/v2",
			Limit:             10,
			Burst:             10,
			VariableMatchType: constant.RateLimitConfigVariableMatchTypePrefix,
			Variable:          constant.RateLimitVariablePath,
			Time:              constant.RateLimitConfigTimeMinute,
		},
	})

	redisStrResult := redis.NewStringResult(string(optsBytes), nil)
	redisStrPrefixResult := redis.NewStringResult(string(optsPrefixBytes), nil)

	testCases := []struct {
		name      string
		req       rateLimitModel.MerchantRateLimitRequest
		setupMock func()
		shouldErr bool
		want      *rateLimitModel.MerchantRateLimitHeaderMetadata
		wantErr   error
	}{
		{
			name: "when rate limit config not found",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redis.NewStringResult("", redis.Nil)).Once()
				mockRateLimitRepo.On("GetMerchantConfigs", constant.ValueCtxMockType(), merchantID).Return(nil, nil).Once()
				mockRedis.On("Set", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything, mock.Anything).Return(redis.NewStatusResult("", nil)).Once()
			},
			shouldErr: false,
			want:      nil,
		},
		{
			name: "when rate limit config found and enforced",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redisStrResult).Once()
				mockLimiter.On("Allow", mock.Anything, merchantCacheKey, limiterCfg).Return(limiterResult, nil).Once()
			},
			shouldErr: false,
			want: &rateLimitModel.MerchantRateLimitHeaderMetadata{
				RateLimitLimit:     config.Limit,
				RateLimitRemaining: 9,
				RateLimitReset:     limiterResult.ResetAfter.Milliseconds(),
			},
		},
		{
			name: "when redis Get fails",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get",
					constant.ValueCtxMockType(),
					merchantCacheConfigKey,
					mock.Anything).Return(
					redis.NewStringResult("", constant.ErrSomeErrorForUnitTest),
				).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrInternal, fmt.Errorf("failed to get configs: %w", constant.ErrSomeErrorForUnitTest)),
		},
		{
			name: "when the config is exact and the path is not matched, skip enforcement",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       "/api/v1/other-path",
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redisStrResult).Once()
			},
			shouldErr: false,
			want:      nil,
		},
		{
			name: "when rate limit enforcement fails, then should return error",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redisStrResult).Once()
				mockLimiter.On("Allow", mock.Anything, merchantCacheKey, limiterCfg).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "when the config is prefix and the path is not matched, skip enforcement",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redisStrPrefixResult).Once()
			},
			shouldErr: false,
			want:      nil,
		},
		{
			name: "when cache is broken, then should return error",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redis.NewStringResult("-", nil)).Once()
			},
			shouldErr: true,
			wantErr:   errors.New("ERROR_INTERNAL | failed to unmarshal rate limit configs: invalid character ' ' in numeric literal"),
		},
		{
			name: "when rate limit exceeded, then should return error",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Path:       path,
				HTTPMethod: "GET",
			},
			setupMock: func() {
				mockRedis.On("Get", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything).Return(redisStrResult).Once()
				mockLimiter.On("Allow", mock.Anything, merchantCacheKey, limiterCfg).Return(&redis_rate.Result{
					Remaining:  0,
					Allowed:    0,
					RetryAfter: limiterResult.ResetAfter,
				}, nil).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrRequestLimitExceeded, fmt.Errorf("request limit exceeded")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			metadata, err := service.ValidateMerchantRateLimit(ctx, tc.req)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, metadata)
		})
	}
}

func TestMerchantRateLimitConfigEvaluation(t *testing.T) {
	config := config.Config{}
	logger, _ := pdkLog.NewZapLogger(pdkLog.Config{})

	ffContentConfig := `
backend-portal-merchant-rate-limit-middleware:
  variations:
    ON: true
    OFF: false
  defaultRule:
    variation: OFF
backend-portal-merchant-default-rate-limit-middleware:
  variations:
    ON: true
    OFF: false
  defaultRule:
    variation: OFF`
	f, err := os.CreateTemp(os.TempDir(), "mock-cfg-*.yaml")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(f.Name())) }()
	defer func() { require.NoError(t, f.Close()) }()

	_, err = f.WriteString(ffContentConfig)
	require.NoError(t, err)

	err = ffclient.Init(ffclient.Config{
		FileFormat: "YAML",
		Retriever: &fileretriever.Retriever{
			Path: f.Name(),
		},
	})
	require.NoError(t, err)
	defer ffclient.Close()

	path := "/api/v1/merchant"

	testCases := []struct {
		name        string
		req         rateLimitModel.MerchantRateLimitRequest
		configs     []rateLimitModel.MerchantRateLimitConfig
		wantErr     bool
		checkLimit  bool
		hasMetadata bool
		err         error
	}{
		{
			name: "ALLOW: Match single configuration with HTTP method",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: uuid.Max.String(),
				Path:       path,
				HTTPMethod: "GET",
			},
			configs: []rateLimitModel.MerchantRateLimitConfig{
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             10,
					Burst:             5,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
					Variable:          constant.RateLimitVariablePath,
					Time:              constant.RateLimitConfigTimeSecond,
					HTTPMethod:        "GET",
				},
			},
			checkLimit:  true,
			hasMetadata: true,
			wantErr:     false,
		},
		{
			name: "ALLOW: Match IP Address single configuration with",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: uuid.Max.String(),
				Path:       path,
				HTTPMethod: "GET",
			},
			configs: []rateLimitModel.MerchantRateLimitConfig{
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             10,
					Burst:             5,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
					Variable:          constant.RateLimitVariableIPAddress,
					Time:              constant.RateLimitConfigTimeSecond,
					HTTPMethod:        "GET",
				},
			},
			checkLimit:  true,
			hasMetadata: true,
			wantErr:     false,
		},
		{
			name: "ALLOW: No configuration matched with HTTP method",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: uuid.Max.String(),
				Path:       path,
				HTTPMethod: "GET",
			},
			configs: []rateLimitModel.MerchantRateLimitConfig{
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             10,
					Burst:             5,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
					Variable:          constant.RateLimitVariablePath,
					Time:              constant.RateLimitConfigTimeSecond,
					HTTPMethod:        "POST",
				},
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             10,
					Burst:             5,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypePrefix,
					Variable:          constant.RateLimitVariablePath,
					Time:              constant.RateLimitConfigTimeSecond,
					HTTPMethod:        "POST",
				},
			},
			checkLimit:  false,
			hasMetadata: false,
			wantErr:     false,
		},
		{
			name: "ALLOW: Match multiple configuration with HTTP method",
			req: rateLimitModel.MerchantRateLimitRequest{
				MerchantID: uuid.Max.String(),
				Path:       path,
				HTTPMethod: "GET",
			},
			configs: []rateLimitModel.MerchantRateLimitConfig{
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             36000,
					Burst:             2500,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypePrefix,
					Variable:          constant.RateLimitVariablePath,
					Time:              constant.RateLimitConfigTimeHour,
					HTTPMethod:        "",
				},
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             600,
					Burst:             50,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
					Variable:          constant.RateLimitVariablePath,
					Time:              constant.RateLimitConfigTimeMinute,
					HTTPMethod:        "GET",
				},
				{
					UUID:              uuid.NewString(),
					VariableValue:     path,
					Limit:             10,
					Burst:             5,
					VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
					Time:              constant.RateLimitConfigTimeSecond,
					HTTPMethod:        "GET",
				},
			},
			checkLimit:  true,
			hasMetadata: true,
			wantErr:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRedis := mockRedisExt.NewIRedisExt(t)
			mockRateLimitRepo := mockRepo.NewIRateLimiterRepository(t)
			mockLimiter := mockRedisExt.NewILimiter(t)

			configB, _ := json.Marshal(tc.configs)
			redisStrConfigResult := redis.NewStringResult(string(configB), nil)

			mockRedis.On("Get", mock.Anything, mock.Anything).Return(redisStrConfigResult).Once()
			rateLimitResult := &redis_rate.Result{
				Limit: redis_rate.Limit{
					Rate: 10,
				},
				Remaining:  9,
				Allowed:    1,
				ResetAfter: time.Second,
			}
			if tc.wantErr {
				rateLimitResult.Allowed = 0
			}
			if tc.checkLimit {
				mockLimiter.On("Allow", mock.Anything, mock.Anything, mock.Anything).Return(rateLimitResult, nil)
			}

			service := New(
				logger,
				mockRedis,
				mockRateLimitRepo,
				WithRedisLimiter(mockLimiter),
				WithConfig(&config))
			metadata, err := service.ValidateMerchantRateLimit(context.Background(), tc.req)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
				return
			}

			if tc.hasMetadata {
				assert.NotZero(t, metadata.RateLimitLimit)
				assert.NotZero(t, metadata.RateLimitRemaining)
				assert.NotZero(t, metadata.RateLimitReset)
			} else {
				assert.Nil(t, metadata)
			}

			assert.NoError(t, err)
		})
	}
}

func TestCacheMerchantRateLimitConfig(t *testing.T) {
	mockLogger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	mockRedis := mockRedisExt.NewIRedisExt(t)
	mockRateLimitRepo := mockRepo.NewIRateLimiterRepository(t)

	traceID := "trace-id"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)
	merchantID := "merchant-id"
	merchantCacheConfigKey := fmt.Sprintf(MerchantRateLimitKeyConfig, merchantID)
	config := rateLimitModel.MerchantRateLimitConfig{
		VariableValue:     "/api/v1/merchant",
		Limit:             10,
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
		Time:              constant.RateLimitConfigTimeMinute,
	}

	service := New(
		mockLogger, mockRedis, mockRateLimitRepo,
	)

	testCases := []struct {
		name      string
		setupMock func()
		shouldErr bool
		want      *[]rateLimitModel.MerchantRateLimitConfig
	}{
		{
			name: "when rate limit config found in repo",
			setupMock: func() {
				mockRateLimitRepo.On("GetMerchantConfigs", constant.ValueCtxMockType(), merchantID).Return(&[]rateLimitModel.MerchantRateLimitConfig{config}, nil).Once()
				mockRedis.On("Set", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything, mock.Anything).Return(redis.NewStatusResult("", nil)).Once()
			},
			shouldErr: false,
			want:      &[]rateLimitModel.MerchantRateLimitConfig{config},
		},
		{
			name: "when rate limit config not found in repo",
			setupMock: func() {
				mockRateLimitRepo.On("GetMerchantConfigs", constant.ValueCtxMockType(), merchantID).Return(nil, nil).Once()
				mockRedis.On("Set", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything, mock.Anything).Return(redis.NewStatusResult("", nil)).Once()
			},
			shouldErr: false,
			want:      &[]rateLimitModel.MerchantRateLimitConfig{},
		},
		{
			name: "when repo returns error",
			setupMock: func() {
				mockRateLimitRepo.On("GetMerchantConfigs", constant.ValueCtxMockType(), merchantID).Return(nil, fmt.Errorf("repo error")).Once()
			},
			shouldErr: true,
			want:      nil,
		},
		{
			name: "when storing cache fails",
			setupMock: func() {
				mockRateLimitRepo.On("GetMerchantConfigs", constant.ValueCtxMockType(), merchantID).Return(&[]rateLimitModel.MerchantRateLimitConfig{config}, nil).Once()
				mockRedis.On("Set", constant.ValueCtxMockType(), merchantCacheConfigKey, mock.Anything, mock.Anything).Return(redis.NewStatusResult("", constant.ErrSomeErrorForUnitTest)).Once()
			},
			shouldErr: true,
			want:      &[]rateLimitModel.MerchantRateLimitConfig{config},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			configs, err := service.CacheMerchantRateLimitConfig(ctx, merchantID)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, configs)
		})
	}
}
func TestEval(t *testing.T) {
	mockLogger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	mockLimiter := mockRedisExt.NewILimiter(t)

	traceID := "trace-id"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)
	key := "rate-limit-key"
	cfg := &rateLimitModel.MerchantRateLimitConfig{
		Limit: 10,
		Time:  constant.RateLimitConfigTimeMinute,
	}

	limitConf := &redisExt.Limit{
		Rate:   cfg.Limit,
		Burst:  cfg.Limit,
		Period: cfg.GetDuration(),
	}

	service := New(
		mockLogger, nil, nil, WithRedisLimiter(mockLimiter),
	)

	testCases := []struct {
		name      string
		setupMock func()
		shouldErr bool
		want      *rateLimitModel.MerchantRateLimitHeaderMetadata
		wantErr   error
	}{
		{
			name: "when rate limit is allowed",
			setupMock: func() {
				mockLimiter.On("Allow", mock.Anything, key, limitConf).Return(&redis_rate.Result{
					Limit:      redis_rate.Limit{Rate: 10},
					Remaining:  9,
					Allowed:    1,
					ResetAfter: cfg.GetDuration() - 10,
				}, nil).Once()
			},
			shouldErr: false,
			want: &rateLimitModel.MerchantRateLimitHeaderMetadata{
				RateLimitLimit:     10,
				RateLimitRemaining: 9,
				RateLimitReset:     (cfg.GetDuration() - 10).Milliseconds(),
			},
		},
		{
			name: "when rate limit is exceeded",
			setupMock: func() {
				mockLimiter.On("Allow", mock.Anything, key, limitConf).Return(&redis_rate.Result{
					Limit:      redis_rate.Limit{Rate: 10},
					Remaining:  0,
					Allowed:    0,
					ResetAfter: cfg.GetDuration() - 10,
				}, nil).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrRequestLimitExceeded, fmt.Errorf("request limit exceeded")),
		},
		{
			name: "when limiter setup fails",
			setupMock: func() {
				mockLimiter.On("Allow", mock.Anything, key, limitConf).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			metadata, err := service.Eval(ctx, key, limitConf)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, metadata)
		})
	}
}
