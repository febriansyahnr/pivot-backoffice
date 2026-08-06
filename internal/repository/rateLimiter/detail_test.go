package rateLimiter_test

import (
	"context"
	"testing"

	"database/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/rateLimiter"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

func TestDetail(t *testing.T) {
	target := ratelimiter.RateLimitConfiguration{}
	tests := []struct {
		name          string
		mockSetup     func(*mysqlMock.IMySqlExt)
		expectedError error
		expectedResp  *ratelimiter.RateLimitConfiguration
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("GetContext", constant.ValueCtxMockType(), constant.PtrRateLimitConfiguration(), constant.StringMockType(), constant.StringMockType()).
					Run(func(args mock.Arguments) {
						dest := args.Get(1).(*ratelimiter.RateLimitConfiguration)
						*dest = target
					}).Return(nil)
			},
			expectedError: nil,
			expectedResp:  &target,
		},
		{
			name: "ERROR: SQL No Rows",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("GetContext", constant.ValueCtxMockType(), constant.PtrRateLimitConfiguration(), constant.StringMockType(), constant.StringMockType()).
					Return(sql.ErrNoRows)
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("GetContext", constant.ValueCtxMockType(), constant.PtrRateLimitConfiguration(), constant.StringMockType(), constant.StringMockType()).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedError: constant.ErrSomeErrorForUnitTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMock.NewIMySqlExt(t)
			mockLogger, _ := pdkLog.NewZapLogger(pdkLog.Config{})

			repo := New(mockMysql, mockLogger)

			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, MerchantRateLimitTable)
			tt.mockSetup(mockMysql)

			resp, err := repo.Detail(ctx, uuid.NewString())
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}
