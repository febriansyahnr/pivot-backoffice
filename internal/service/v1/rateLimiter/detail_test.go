package ratelimiter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockAnything = mock.Anything

func TestRateLimiterService_Detail(t *testing.T) {
	logger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	ctx := context.Background()

	merchantID := uuid.New().String()
	configUUID := uuid.New().String()
	otherMerchantID := uuid.New().String()

	mockConfig := &ratelimiter.RateLimitConfiguration{
		UUID:              configUUID,
		MerchantID:        merchantID,
		Limit:             10,
		Burst:             5,
		Order:             1,
		Time:              "MINUTE",
		Variable:          "IP",
		VariableValue:     "127.0.0.1",
		VariableMatchType: "EXACT",
		HTTPMethod:        "POST",
		Status:            constant.StatusActive,
		Description:       "Test rate limit",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	tests := []struct {
		name        string
		merchantID  string
		configUUID  string
		mockSetup   func(mockRateLimitRepo *mockRepo.IRateLimiterRepository)
		wantErr     bool
		expectedErr string
		validateResult func(t *testing.T, result *ratelimiter.RateLimitConfiguration)
	}{
		{
			name:       "SUCCESS: Get rate limit configuration detail",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(mockConfig, nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ratelimiter.RateLimitConfiguration) {
				assert.NotNil(t, result)
				assert.Equal(t, configUUID, result.UUID)
				assert.Equal(t, merchantID, result.MerchantID)
				assert.Equal(t, 10, result.Limit)
				assert.Equal(t, 5, result.Burst)
				assert.Equal(t, 1, result.Order)
				assert.Equal(t, "MINUTE", result.Time)
				assert.Equal(t, "IP", result.Variable)
				assert.Equal(t, "127.0.0.1", result.VariableValue)
				assert.Equal(t, "EXACT", result.VariableMatchType)
				assert.Equal(t, "POST", result.HTTPMethod)
				assert.Equal(t, constant.StatusActive, result.Status)
				assert.Equal(t, "Test rate limit", result.Description)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			},
		},
		{
			name:       "ERROR: Repository Detail fails",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(nil, errors.New("database error")).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrGetRateLimiterConfigurationDetail.Error(),
		},
		{
			name:       "ERROR: Configuration not found",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(nil, nil).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrRateLimiterConfigurationNotFound.Error(),
		},
		{
			name:       "ERROR: Invalid merchant ID - configuration belongs to different merchant",
			merchantID: otherMerchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(mockConfig, nil).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrInvalidRateLimiterConfigurationID.Error(),
		},
		{
			name:       "SUCCESS: Get configuration with different settings",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				config := &ratelimiter.RateLimitConfiguration{
					UUID:              configUUID,
					MerchantID:        merchantID,
					Limit:             100,
					Burst:             20,
					Order:             5,
					Time:              "HOUR",
					Variable:          "PATH",
					VariableValue:     "/api/payment",
					VariableMatchType: "PREFIX",
					HTTPMethod:        "GET",
					Status:            constant.StatusActive,
					Description:       "API rate limit",
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
				}
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(config, nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ratelimiter.RateLimitConfiguration) {
				assert.NotNil(t, result)
				assert.Equal(t, configUUID, result.UUID)
				assert.Equal(t, merchantID, result.MerchantID)
				assert.Equal(t, 100, result.Limit)
				assert.Equal(t, 20, result.Burst)
				assert.Equal(t, 5, result.Order)
				assert.Equal(t, "HOUR", result.Time)
				assert.Equal(t, "PATH", result.Variable)
				assert.Equal(t, "/api/payment", result.VariableValue)
				assert.Equal(t, "PREFIX", result.VariableMatchType)
				assert.Equal(t, "GET", result.HTTPMethod)
				assert.Equal(t, "API rate limit", result.Description)
			},
		},
		{
			name:       "SUCCESS: Get inactive configuration",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				config := &ratelimiter.RateLimitConfiguration{
					UUID:              configUUID,
					MerchantID:        merchantID,
					Limit:             50,
					Burst:             10,
					Order:             3,
					Time:              "DAILY",
					Variable:          "IP",
					VariableValue:     "192.168.1.0",
					VariableMatchType: "CONTAINS",
					HTTPMethod:        "PUT",
					Status:            constant.StatusInactive,
					Description:       "Inactive rate limit",
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
				}
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(config, nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ratelimiter.RateLimitConfiguration) {
				assert.NotNil(t, result)
				assert.Equal(t, constant.StatusInactive, result.Status)
				assert.Equal(t, "DAILY", result.Time)
				assert.Equal(t, "CONTAINS", result.VariableMatchType)
				assert.Equal(t, "PUT", result.HTTPMethod)
			},
		},
		{
			name:       "SUCCESS: Get configuration with maximum values",
			merchantID: merchantID,
			configUUID: configUUID,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				config := &ratelimiter.RateLimitConfiguration{
					UUID:              configUUID,
					MerchantID:        merchantID,
					Limit:             1000,
					Burst:             200,
					Order:             999,
					Time:              "SECOND",
					Variable:          "PATH",
					VariableValue:     "/api/v1/payment/process",
					VariableMatchType: "EXACT",
					HTTPMethod:        "DELETE",
					Status:            constant.StatusActive,
					Description:       "High volume API endpoint",
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
				}
				mockRateLimitRepo.On("Detail", mockAnything, configUUID).Return(config, nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ratelimiter.RateLimitConfiguration) {
				assert.NotNil(t, result)
				assert.Equal(t, 1000, result.Limit)
				assert.Equal(t, 200, result.Burst)
				assert.Equal(t, 999, result.Order)
				assert.Equal(t, "SECOND", result.Time)
				assert.Equal(t, "DELETE", result.HTTPMethod)
				assert.Equal(t, "High volume API endpoint", result.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRateLimitRepo := mockRepo.NewIRateLimiterRepository(t)

			service := &rateLimiterService{
				logger:          logger,
				rateLimiterRepo: mockRateLimitRepo,
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mockRateLimitRepo)
			}

			result, err := service.Detail(ctx, tt.merchantID, tt.configUUID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			mockRateLimitRepo.AssertExpectations(t)
		})
	}
}