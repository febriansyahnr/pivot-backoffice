package ratelimiter

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)


func TestRateLimiterService_List(t *testing.T) {
	logger, _ := pdkLog.NewZapLogger(pdkLog.Config{})
	ctx := context.Background()

	merchantID := uuid.New().String()

	validRequest := &ratelimiter.MerchantRateLimitRequest{
		MerchantID: merchantID,
		Page:       1,
		PageSize:   10,
	}

	mockConfigs := []*ratelimiter.RateLimitConfiguration{
		{
			UUID:              uuid.New().String(),
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
		},
		{
			UUID:              uuid.New().String(),
			MerchantID:        merchantID,
			Limit:             100,
			Burst:             20,
			Order:             2,
			Time:              "HOUR",
			Variable:          "PATH",
			VariableValue:     "/api/payment",
			VariableMatchType: "PREFIX",
			HTTPMethod:        "GET",
			Status:            constant.StatusActive,
			Description:       "API rate limit",
		},
	}

	tests := []struct {
		name        string
		request     *ratelimiter.MerchantRateLimitRequest
		mockSetup   func(mockRateLimitRepo *mockRepo.IRateLimiterRepository)
		wantErr     bool
		expectedErr string
		validateResult func(t *testing.T, result *commonModel.PaginationResponse)
	}{
		{
			name:    "SUCCESS: Get list with results",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("List", mock.Anything, validRequest).Return(mockConfigs, int64(2), nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, mockConfigs, result.Data)
				assert.Equal(t, int64(1), result.Meta.Page)
				assert.Equal(t, int64(10), result.Meta.PerPage)
				assert.Equal(t, int64(2), result.Meta.TotalItems)
				assert.Equal(t, int64(1), result.Meta.TotalPages)
			},
		},
		{
			name:    "SUCCESS: Get empty list",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("List", mock.Anything, validRequest).Return([]*ratelimiter.RateLimitConfiguration{}, int64(0), nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Empty(t, result.Data)
				assert.Equal(t, int64(0), result.Meta.TotalItems)
			},
		},
		{
			name: "SUCCESS: Get list with pagination",
			request: &ratelimiter.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Page:       2,
				PageSize:   5,
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				req := &ratelimiter.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Page:       2,
					PageSize:   5,
				}
				mockRateLimitRepo.On("List", mock.Anything, req).Return(mockConfigs[1:], int64(10), nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Len(t, result.Data, 1)
				assert.Equal(t, int64(2), result.Meta.Page)
				assert.Equal(t, int64(5), result.Meta.PerPage)
				assert.Equal(t, int64(10), result.Meta.TotalItems)
				assert.Equal(t, int64(2), result.Meta.TotalPages)
			},
		},
		{
			name: "SUCCESS: Get list with filters",
			request: &ratelimiter.MerchantRateLimitRequest{
				MerchantID:    merchantID,
				Variable:      "IP",
				VariableValue: "127.0.0.1",
				HTTPMethod:    "POST",
				Page:          1,
				PageSize:      10,
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				req := &ratelimiter.MerchantRateLimitRequest{
					MerchantID:    merchantID,
					Variable:      "IP",
					VariableValue: "127.0.0.1",
					HTTPMethod:    "POST",
					Page:          1,
					PageSize:      10,
				}
				mockRateLimitRepo.On("List", mock.Anything, req).Return(mockConfigs[:1], int64(1), nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Len(t, result.Data, 1)
				configs := result.Data.([]*ratelimiter.RateLimitConfiguration)
				assert.Equal(t, "IP", configs[0].Variable)
				assert.Equal(t, "127.0.0.1", configs[0].VariableValue)
				assert.Equal(t, "POST", configs[0].HTTPMethod)
			},
		},
		{
			name:    "ERROR: Repository List fails",
			request: validRequest,
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				mockRateLimitRepo.On("List", mock.Anything, validRequest).Return(nil, int64(0), errors.New("database error")).Once()
			},
			wantErr:     true,
			expectedErr: constant.ErrGetRateLimiterConfigurationList.Error(),
		},
		{
			name: "SUCCESS: Get list with status filter",
			request: &ratelimiter.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Status:     constant.StatusActive,
				Page:       1,
				PageSize:   10,
			},
			mockSetup: func(mockRateLimitRepo *mockRepo.IRateLimiterRepository) {
				req := &ratelimiter.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Status:     constant.StatusActive,
					Page:       1,
					PageSize:   10,
				}
				mockRateLimitRepo.On("List", mock.Anything, req).Return(mockConfigs, int64(2), nil).Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Len(t, result.Data, 2)
				configs := result.Data.([]*ratelimiter.RateLimitConfiguration)
				for _, config := range configs {
					assert.Equal(t, constant.StatusActive, config.Status)
				}
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

			result, err := service.List(ctx, tt.request)

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