package callbackRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	merchantID := uuid.NewString()
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows)
		filter    *callbackModel.GetListCallbackFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &callbackModel.GetListCallbackFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  &callbackModel.GetListCallbackFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with merchantID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: &callbackModel.GetListCallbackFilterRequest{
				MerchantID: &merchantID,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("invalid table name"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)

			},
			filter:  &callbackModel.GetListCallbackFilterRequest{},
			wantErr: true,
		},
		{
			name: "Get List on error retrieving data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("invalid scan data"))

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)

			},
			filter:  &callbackModel.GetListCallbackFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysqlRows := mysqlMocks.NewIMySqlRows(t)
			tc.mockSetup(mockMysql, mockMysqlRows)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
