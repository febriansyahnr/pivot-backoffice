package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSubMerchantByParentID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *merchantModel.SubMerchantListFilter
		wantErr   bool
	}{
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
				).Return(errors.New("database error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &merchantModel.SubMerchantListFilter{},
			wantErr: true,
		},
		{
			name: "SUCCESS: List Sub Merchants",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					merchantsPtr := args.Get(1).(*[]*merchantModel.Merchant)
					*merchantsPtr = []*merchantModel.Merchant{
						{UUID: "merchant-uuid-1", Name: "Merchant 1"},
					}
				})

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &merchantModel.SubMerchantListFilter{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get sub merchant list without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter:  &merchantModel.SubMerchantListFilter{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get list of sub merchants with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				ParentId: "parent-id",
				MID:      "mid-123",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by Name",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				Name: "Test Merchant",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by ShortName",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				ShortName: "TM",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by Email",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by Keywords",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				Keywords: "search term",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by Status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				Status: "ACTIVE",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Filter by Date Range",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &merchantModel.SubMerchantListFilter{
				StartCreatedAt: &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
				EndCreatedAt:   &[]time.Time{time.Now()}[0],
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "merchants")
			_, err := repo.ListSubMerchantByParentID(ctx, tc.filter, 0, 20)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetSubMerchantsByParentID(t *testing.T) {
	testCases := []struct {
		Name      string
		MockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Get sub merchants by parent ID",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: No sub merchants found (sql.ErrNoRows)",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Database error",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			WantErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			test.MockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetSubMerchantsByParentID(context.Background(), "parent-merchant-id-123")

			if test.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
