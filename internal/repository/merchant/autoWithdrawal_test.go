package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListOfMerchantsWithActiveAutoWithdrawalStatus(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db, nil)

	result := []merchant.MerchantWithActiveAutoWithdrawalStatus{
		{
			MerchantId:           "8a8a8405-5abd-48ea-a47f-b4bd3b03f26c",
			MerchantName:         "PT Dummy",
			AccountName:          "PAYMENT",
			BeneficiaryBankCode:  "002",
			BeneficiaryAccountNo: "111111111111",
		},
	}
	resultMockType := mock.AnythingOfType("*[]merchant.MerchantWithActiveAutoWithdrawalStatus")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []merchant.MerchantWithActiveAutoWithdrawalStatus
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.MerchantWithActiveAutoWithdrawalStatus)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListOfMerchantsWithActiveAutoWithdrawalStatus(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetListOfMerchantsToForceTheAutoWithdrawalProcess(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db, nil)

	result := []merchant.MerchantWithdrawalDetails{
		{MerchantId: "5db6cb4e-cf48-4e31-9a42-0384eea81e1c"},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []merchant.MerchantWithdrawalDetails
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: []merchant.MerchantWithdrawalDetails{},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.MerchantWithdrawalDetails)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListOfMerchantsToForceTheAutoWithdrawalProcess(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
