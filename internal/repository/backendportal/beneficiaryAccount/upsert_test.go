package beneficiaryAccountRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func TestUpsert(t *testing.T) {
	account := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-merchant-merchant",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryAccountName: "testing",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing bank",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		account   *beneficiaryAccountModel.BeneficiaryAccount
		wantErr   bool
	}{
		{
			name:    "Valid Account",
			account: account,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"NamedExecContext",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
					).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Upsert(context.Background(), tc.account)

			if (err != nil) != tc.wantErr {
				t.Errorf("BeneficiaryAccount.Upsert() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
