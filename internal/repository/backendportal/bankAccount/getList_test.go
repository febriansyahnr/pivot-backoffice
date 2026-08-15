package bankAccountRepository_test

import (
	"context"
	"database/sql"
	"testing"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/bankAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/bankAccount"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListBankAccount(t *testing.T) {
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
	resultMockType := mock.AnythingOfType("*[]bankAccount.BankAccountResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []bankAccount.BankAccountResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), merchantId,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]bankAccount.BankAccountResponse) = []bankAccount.BankAccountResponse{result}
				}).Return(nil)
			},
			wantResult: []bankAccount.BankAccountResponse{result},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListBankAccount(context.Background(), merchantId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
