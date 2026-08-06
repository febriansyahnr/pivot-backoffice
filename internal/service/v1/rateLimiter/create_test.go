package ratelimiter

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockRedisExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterService_Create(t *testing.T) {
	logger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	ctx := context.Background()

	merchantID := uuid.New().String()

	validRequest := &ratelimiter.CreateRateLimitConfiguration{
		MerchantID:        merchantID,
		Limit:             10,
		Burst:             5,
		Order:             1,
		Time:              "MINUTE",
		Variable:          "IP",
		VariableValue:     "127.0.0.1",
		VariableMatchType: "EXACT",
		HTTPMethod:        "POST",
		Description:       "Test rate limit",
	}

	tests := []struct {
		name        string
		request     *ratelimiter.CreateRateLimitConfiguration
		mockSetup   func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "SUCCESS: Create rate limit configuration",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				// Mock repository create
				mockRateLimitRepo.On("Create", mock.Anything, mock.MatchedBy(func(config *ratelimiter.RateLimitConfiguration) bool {
					return config.MerchantID == merchantID &&
						config.Limit == 10 &&
						config.Burst == 5 &&
						config.Order == 1 &&
						config.Time == "MINUTE" &&
						config.Variable == "IP" &&
						config.VariableValue == "127.0.0.1" &&
						config.VariableMatchType == "EXACT" &&
						config.HTTPMethod == "POST" &&
						config.Status == constant.StatusActive &&
						config.Description == "Test rate limit"
				})).Return(nil).Once()

				// Mock cache operations for CacheMerchantRateLimitConfig
				configs := &[]ratelimiter.MerchantRateLimitConfig{
					{
						UUID:              uuid.New().String(),
						Variable:          "IP",
						VariableValue:     "127.0.0.1",
						Limit:             10,
						Burst:             5,
						VariableMatchType: "EXACT",
						HTTPMethod:        "POST",
						Time:              "MINUTE",
					},
				}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()

				// Mock Redis Set operation
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(nil)
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Repository Create fails",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrCreateRateLimiterConfiguration.Error(),
		},
		{
			name:    "ERROR: CacheMerchantRateLimitConfig fails - GetByMerchantID error",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")).
					Return(nil).Once()

				// Simulate cache failure by making GetMerchantConfigs fail
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(nil, errors.New("database error")).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrCreateRateLimiterConfiguration.Error(),
		},
		{
			name:    "ERROR: CacheMerchantRateLimitConfig fails - Redis Set error",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")).
					Return(nil).Once()

				configs := &[]ratelimiter.MerchantRateLimitConfig{}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()

				// Simulate Redis Set failure
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(errors.New("redis error"))
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrCreateRateLimiterConfiguration.Error(),
		},
		{
			name: "SUCCESS: Create with minimum required fields",
			request: &ratelimiter.CreateRateLimitConfiguration{
				MerchantID:        merchantID,
				Limit:             1,
				Burst:             0,
				Order:             0,
				Time:              "SECOND",
				Variable:          "PATH",
				VariableValue:     "/api/payment",
				VariableMatchType: "PREFIX",
				HTTPMethod:        "GET",
				Description:       "",
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.MatchedBy(func(config *ratelimiter.RateLimitConfiguration) bool {
					return config.MerchantID == merchantID &&
						config.Limit == 1 &&
						config.Burst == 0 &&
						config.Order == 0 &&
						config.Time == "SECOND" &&
						config.Variable == "PATH" &&
						config.VariableValue == "/api/payment" &&
						config.VariableMatchType == "PREFIX" &&
						config.HTTPMethod == "GET" &&
						config.Status == constant.StatusActive &&
						config.Description == ""
				})).Return(nil).Once()

				configs := &[]ratelimiter.MerchantRateLimitConfig{}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()
				
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(nil)
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Create with maximum values",
			request: &ratelimiter.CreateRateLimitConfiguration{
				MerchantID:        merchantID,
				Limit:             1000,
				Burst:             200,
				Order:             999,
				Time:              "DAILY",
				Variable:          "IP",
				VariableValue:     "192.168.1.0/24",
				VariableMatchType: "CONTAINS",
				HTTPMethod:        "DELETE",
				Description:       "High limit configuration for premium merchants",
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.MatchedBy(func(config *ratelimiter.RateLimitConfiguration) bool {
					return config.MerchantID == merchantID &&
						config.Limit == 1000 &&
						config.Burst == 200 &&
						config.Order == 999 &&
						config.Time == "DAILY" &&
						config.Variable == "IP" &&
						config.VariableValue == "192.168.1.0/24" &&
						config.VariableMatchType == "CONTAINS" &&
						config.HTTPMethod == "DELETE" &&
						config.Status == constant.StatusActive
				})).Return(nil).Once()

				configs := &[]ratelimiter.MerchantRateLimitConfig{}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()
				
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(nil)
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Create with HOUR time setting",
			request: &ratelimiter.CreateRateLimitConfiguration{
				MerchantID:        merchantID,
				Limit:             100,
				Burst:             20,
				Order:             5,
				Time:              "HOUR",
				Variable:          "PATH",
				VariableValue:     "/api/v1/payment",
				VariableMatchType: "EXACT",
				HTTPMethod:        "PUT",
				Description:       "Hourly limit for payment API",
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")).
					Return(nil).Once()

				configs := &[]ratelimiter.MerchantRateLimitConfig{}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()
				
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(nil)
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Create with PATCH method",
			request: &ratelimiter.CreateRateLimitConfiguration{
				MerchantID:        merchantID,
				Limit:             50,
				Burst:             10,
				Order:             3,
				Time:              "MINUTE",
				Variable:          "IP",
				VariableValue:     "10.0.0.1",
				VariableMatchType: "EXACT",
				HTTPMethod:        "PATCH",
				Description:       "PATCH method rate limit",
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository, mockRedis *mockRedisExt.IRedisExt, service *rateLimiterService) {
				mockRateLimitRepo.On("Create", mock.Anything, mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")).
					Return(nil).Once()

				configs := &[]ratelimiter.MerchantRateLimitConfig{}
				mockRateLimitRepo.On("GetMerchantConfigs", mock.Anything, merchantID).Return(configs, nil).Once()
				
				mockStatusCmd := &redis.StatusCmd{}
				mockStatusCmd.SetErr(nil)
				mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(mockStatusCmd).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRateLimitRepo := mockRepo.NewIRateLimiterRepository(t)
			mockRedis := mockRedisExt.NewIRedisExt(t)
			
			service := &rateLimiterService{
				logger:          logger,
				redis:           mockRedis,
				rateLimiterRepo: mockRateLimitRepo,
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mockRateLimitRepo, mockRedis, service)
			}

			result, err := service.Create(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.UUID)
				assert.Equal(t, tt.request.MerchantID, result.MerchantID)
				assert.Equal(t, tt.request.Limit, result.Limit)
				assert.Equal(t, tt.request.Burst, result.Burst)
				assert.Equal(t, tt.request.Order, result.Order)
				assert.Equal(t, tt.request.Time, result.Time)
				assert.Equal(t, tt.request.Variable, result.Variable)
				assert.Equal(t, tt.request.VariableValue, result.VariableValue)
				assert.Equal(t, tt.request.VariableMatchType, result.VariableMatchType)
				assert.Equal(t, tt.request.HTTPMethod, result.HTTPMethod)
				assert.Equal(t, tt.request.Description, result.Description)
				assert.Equal(t, constant.StatusActive, result.Status)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			}

			mockRateLimitRepo.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}