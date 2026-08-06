package bankAccountRepository

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {

	validData := &bankAccount.BankAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "12345",
		BeneficiaryAccountName: "Kai",
		BeneficiaryBankCode:    "001",
		BeneficiaryBankName:    "BCA",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	dataMockType := mock.AnythingOfType("*bankAccount.BankAccount")

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(db *mysqlMocks.IMySqlExt)
	}{
		{
			desc:    "success Update data",
			wantErr: false,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.
					On(
						"NamedExecContext",
						mock.Anything,
						c.StringMockType(),
						dataMockType).
					Return(true, nil)
			},
		},
		{
			desc:    "error Update data",
			wantErr: true,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.
					On(
						"NamedExecContext",
						mock.Anything,
						c.StringMockType(),
						dataMockType).
					Return(false, assert.AnError)
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			db := mysqlMocks.NewIMySqlExt(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(db)

			repo := New(db, logger)
			err := repo.Update(context.Background(), validData)
			if (err != nil) != tt.wantErr {
				t.Errorf("BankAccount.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
