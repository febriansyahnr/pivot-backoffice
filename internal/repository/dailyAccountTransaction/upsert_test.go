package dailyAccountTransactionRepository_test

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/dailyAccountTransaction"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestUpsert(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Upsert(ctx, &dailyAccountTransactionModel.DailyAccountTransaction{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
