package rateLimiter

import (
	"context"
	"database/sql"
	"testing"

	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/rateLimiter"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantConfigs(t *testing.T) {
	ctx := context.Background()
	merchantID := "test-merchant-id"
	target := &[]ratelimiterModel.MerchantRateLimitConfig{
		{
			VariableValue:     "test-variable",
			Limit:             100,
			VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
			Time:              constant.RateLimitConfigTimeMinute,
		},
	}

	tests := []struct {
		name           string
		mockSetup      func(*mysqlMock.IMySqlExt)
		expectedResult *[]ratelimiterModel.MerchantRateLimitConfig
		expectedError  error
	}{
		{
			name: "when no error occurs, then return the rate limit configs",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, merchantID).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]ratelimiterModel.MerchantRateLimitConfig)
					*dest = *target
				}).Return(nil).Once()
			},
			expectedResult: target,
			expectedError:  nil,
		},
		{
			name: "merchant does not have any rate limit configs",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, merchantID).Return(nil).Once()
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name: "error",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, merchantID).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedResult: nil,
			expectedError:  constant.ErrSomeErrorForUnitTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMock.NewIMySqlExt(t)
			mockLogger, _ := pdkLog.NewZapLogger(pdkLog.Config{})

			repo := New(mockMysql, mockLogger)

			tt.mockSetup(mockMysql)
			configs, err := repo.GetMerchantConfigs(ctx, merchantID)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, configs)
		})
	}
}
func TestList(t *testing.T) {
	ctx := context.Background()
	merchantID := "test-merchant-id"
	rateLimitConfig := &[]*ratelimiterModel.RateLimitConfiguration{
		{
			Variable:          "ip",
			VariableValue:     "192.168.1.1",
			VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
			Limit:             100,
			Time:              constant.RateLimitConfigTimeMinute,
			Order:             1,
		},
	}

	tests := []struct {
		name           string
		request        *ratelimiterModel.MerchantRateLimitRequest
		mockSetup      func(*mysqlMock.IMySqlExt)
		expectedResult []*ratelimiterModel.RateLimitConfiguration
		expectedTotal  int64
		expectedError  error
	}{
		{
			name: "when no error occurs, then return the rate limit configurations",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Status:     "active",
				Variable:   "ip",
				Page:       1,
				PageSize:   10,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						dest := args.Get(1).(*int64)
						*dest = 1
					}).
					Return(nil).Once()
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*ratelimiterModel.RateLimitConfiguration)
					*dest = *rateLimitConfig
				}).Return(nil).Once()
			},
			expectedResult: *rateLimitConfig,
			expectedTotal:  1,
			expectedError:  nil,
		},
		{
			name: "when no filters are applied, return all configurations",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything).
					Run(func(args mock.Arguments) {
						dest := args.Get(1).(*int64)
						*dest = 1
					}).Return(nil).Once()
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*ratelimiterModel.RateLimitConfiguration)
					*dest = *rateLimitConfig
				}).Return(nil).Once()
			},
			expectedResult: *rateLimitConfig,
			expectedTotal:  1,
			expectedError:  nil,
		},
		{
			name: "when request has no pagination, return results",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						dest := args.Get(1).(*int64)
						*dest = 1
					}).Return(nil).Once()
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*ratelimiterModel.RateLimitConfiguration)
					*dest = *rateLimitConfig
				}).Return(nil).Once()
			},
			expectedResult: *rateLimitConfig,
			expectedTotal:  1,
			expectedError:  nil,
		},
		{
			name: "when GetContext returns error, return error",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Page:       1,
				PageSize:   10,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			},
			expectedResult: nil,
			expectedTotal:  0,
			expectedError:  constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when SelectContext returns error, return error",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Page:       1,
				PageSize:   10,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResult: nil,
			expectedTotal:  0,
			expectedError:  constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when no results found, return empty list",
			request: &ratelimiterModel.MerchantRateLimitRequest{
				MerchantID: merchantID,
				Page:       1,
				PageSize:   10,
			},
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("Rebind", mock.Anything).Return(mock.Anything).Times(2)
				mockMysql.On("GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*int64"), mock.Anything, mock.Anything).Return(nil).Once()
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*ratelimiterModel.RateLimitConfiguration)
					*dest = []*ratelimiterModel.RateLimitConfiguration{}
				}).Return(sql.ErrNoRows).Once()
			},
			expectedResult: []*ratelimiterModel.RateLimitConfiguration{},
			expectedTotal:  0,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMock.NewIMySqlExt(t)
			mockLogger, _ := pdkLog.NewZapLogger(pdkLog.Config{})

			repo := New(mockMysql, mockLogger)

			tt.mockSetup(mockMysql)
			configs, total, err := repo.List(ctx, tt.request)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, configs)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}
