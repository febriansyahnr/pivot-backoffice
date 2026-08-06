package ipWhitelistRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update Configuration",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*ipwhitelistModel.IPWhitelistConfiguration"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*ipwhitelistModel.IPWhitelistConfiguration"),
				).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*ipwhitelistModel.IPWhitelistConfiguration"),
				).Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)
			err := repo.Update(ctx, &ipwhitelistModel.IPWhitelistConfiguration{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
