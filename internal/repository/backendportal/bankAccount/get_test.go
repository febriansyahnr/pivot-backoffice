package bankAccountRepository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/bankAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/bankAccount"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBankAccountValidation(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	repo := New(db, logger)

	merchantId := uuid.NewString()
	result := bankAccount.BankAccountResponse{
		BeneficiaryBankCode:    "002",
		BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
		BeneficiaryAccountNo:   "999966660002",
		BeneficiaryAccountName: "John",
	}
	resultMockType := mock.AnythingOfType("*bankAccount.BankAccountResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *bankAccount.BankAccountResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId, result.BeneficiaryBankCode, result.BeneficiaryAccountNo,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest, // NOSONAR
			wantResult: &bankAccount.BankAccountResponse{},
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId, result.BeneficiaryBankCode, result.BeneficiaryAccountNo,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId, result.BeneficiaryBankCode, result.BeneficiaryAccountNo,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*bankAccount.BankAccountResponse)) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetBankAccountValidation(context.Background(), merchantId, result.BeneficiaryBankCode, result.BeneficiaryAccountNo)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetByMerchantID(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	timeNow := time.Now()

	repo := New(db, logger)

	merchantId := uuid.NewString()
	result := bankAccount.BankAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             merchantId,
		BeneficiaryBankCode:    "002",
		BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
		BeneficiaryAccountNo:   "999966660002",
		BeneficiaryAccountName: "John",
		CreatedBy:              "John",
		CreatedAt:              timeNow,
		UpdatedBy:              "John",
		UpdatedAt:              timeNow,
	}
	resultMockType := mock.AnythingOfType("*bankAccount.BankAccount")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *bankAccount.BankAccount
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest, // NOSONAR
			wantResult: &bankAccount.BankAccount{},
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*bankAccount.BankAccount)) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetByMerchantID(context.Background(), merchantId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestBankAccountHasBeenPrepared(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult bool
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrBoolMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrBoolMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*bool)) = true
				}).Return(nil)
			},
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.BankAccountHasBeenPrepared(context.Background(), "12345")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
