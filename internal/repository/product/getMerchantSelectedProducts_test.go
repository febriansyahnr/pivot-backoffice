package productRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantSelectedProducts(t *testing.T) {

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant selected products",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*product.MerchantWithProductName"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure get from Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*product.MerchantWithProductName"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*product.MerchantWithProductName"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetMerchantSelectedProducts(context.Background(), uuid.NewString())

			if (err != nil) != tc.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestGetMerchantActiveProducts(t *testing.T) {

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant selected products",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure get from Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetMerchantActiveProducts(context.Background(), uuid.NewString())

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}
