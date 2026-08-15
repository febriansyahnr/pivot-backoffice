package account_repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByReferenceIDAndUsecase(t *testing.T) {

	testCases := []struct {
		Name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			Name: "SUCCESS: Get By Reference ID and Usecase",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					constant.StringMockType(),
					constant.UuidMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			Name: "ERROR: No Rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					constant.StringMockType(),
					constant.UuidMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows).Once()
			},
			wantErr: false,
		},

		{
			Name: "ERROR: Get Errors",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					constant.StringMockType(),
					constant.UuidMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("errors"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, TableName)
			_, err := repo.GetByReferenceIDAndUsecase(ctx, uuid.New(), "usecase", "userType")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
