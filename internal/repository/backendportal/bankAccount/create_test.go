package bankAccountRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	testNow := time.Now()

	dataMockType := mock.AnythingOfType("*bankAccount.BankAccount")

	tests := []struct {
		name      string
		input     *bankAccount.BankAccount
		setupMock func(mysqlMock *mySqlExtMock.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS", // NOSONAR
			input: &bankAccount.BankAccount{
				UUID:                   "uuid-uuid-uuid",
				MerchantID:             "merchant-id",
				BeneficiaryAccountNo:   "12341234",
				BeneficiaryAccountName: "testing",
				BeneficiaryBankCode:    "1234",
				BeneficiaryBankName:    "testing bank",
				CreatedAt:              testNow,
				UpdatedAt:              testNow,
			},
			setupMock: func(mysqlMock *mySqlExtMock.IMySqlExt) {
				mysqlMock.
					On(
						"NamedExecContext",
						mock.AnythingOfType(c.MockTypeValueContextReference),
						c.StringMockType(),
						dataMockType,
					).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database", // NOSONAR
			setupMock: func(mysqlMock *mySqlExtMock.IMySqlExt) {
				mysqlMock.
					On(
						"NamedExecContext",
						mock.AnythingOfType(c.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType(c.MockTypeBankAccountReference),
					).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected", // NOSONAR
			setupMock: func(mysqlMock *mySqlExtMock.IMySqlExt) {
				mysqlMock.
					On(
						"NamedExecContext",
						mock.AnythingOfType(c.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType(c.MockTypeBankAccountReference),
					).Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mySqlExtMock.NewIMySqlExt(t)
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

			tt.setupMock(db)

			repo := New(db, logger)

			err := repo.Create(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountInquiries.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
