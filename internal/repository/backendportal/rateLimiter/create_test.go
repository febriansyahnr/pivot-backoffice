package rateLimiter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/rateLimiter"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/rateLimiter"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func TestCreate(t *testing.T) {

	tests := []struct {
		name          string
		mockSetup     func(*mysqlMock.IMySqlExt)
		expectedError error
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrRateLimitConfiguration(),
				).Return(true, nil)
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mockMysql *mysqlMock.IMySqlExt) {
				mockMysql.On("NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrRateLimitConfiguration(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
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

			err := repo.Create(ctx, &ratelimiter.RateLimitConfiguration{})
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
