package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListOfMerchantsWhoHaveSubMerchant(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchants := []merchant.MerchantWithSubMerchantList{
		{
			ID:              "1",
			RawSubMerchants: []byte(`["A","B","C"]`),
			SubMerchants:    []string{"A", "B", "C"},
		},
	}
	ptrDataMockType := mock.AnythingOfType("*[]merchant.MerchantWithSubMerchantList")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []merchant.MerchantWithSubMerchantList
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrDataMockType, c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrDataMockType, c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrDataMockType, c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.MerchantWithSubMerchantList)) = []merchant.MerchantWithSubMerchantList{
						{
							ID: "1", RawSubMerchants: []byte(`["A","B","C"]`),
						},
					}
				}).Return(nil)
			},
			wantResult: merchants,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListOfMerchantsWhoHaveSubMerchant(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantFeeListForBalanceDeduction(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchants := []merchant.MerchantFeeForBalanceDeduction{
		{
			MerchantId: "12345", Reference: "TEST",
		},
	}
	ptrResultDataMockType := mock.AnythingOfType("*[]merchant.MerchantFeeForBalanceDeduction")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []merchant.MerchantFeeForBalanceDeduction
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrResultDataMockType, c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrResultDataMockType, c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrResultDataMockType, c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.MerchantFeeForBalanceDeduction)) = []merchant.MerchantFeeForBalanceDeduction{
						{
							MerchantId: "12345", Reference: "TEST",
						},
					}
				}).Return(nil)
			},
			wantResult: merchants,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetMerchantFeeListForBalanceDeduction(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
