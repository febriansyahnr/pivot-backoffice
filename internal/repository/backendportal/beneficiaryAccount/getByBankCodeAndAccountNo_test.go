package beneficiaryAccountRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeneficiaryAccountRepository_GetByBankCodeAndAccountNo(t *testing.T) {
	account := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryAccountName: "testing",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing bank",
		Metadata: types.NullJSONText{
			Valid:    true,
			JSONText: []byte{},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		bankCode   string
		accountNo  string
		expected   *beneficiaryAccountModel.BeneficiaryAccount
		wantErr    bool
	}{
		{
			name: "SUCCESS: Find Account by Bank Code, Account No and Merchant ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*beneficiaryAccountModel.BeneficiaryAccount)
					*userPtr = *account
				})
			},
			merchantId: account.MerchantID,
			bankCode:   account.BeneficiaryBankCode,
			accountNo:  account.BeneficiaryAccountNo,
			expected:   account,
			wantErr:    false,
		},
		{
			name: "ERROR: Account Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(sql.ErrNoRows)
			},
			merchantId: account.MerchantID,
			bankCode:   account.BeneficiaryBankCode,
			accountNo:  account.BeneficiaryAccountNo,
			expected:   nil,
			wantErr:    false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(errors.New("database error"))
			},
			merchantId: account.MerchantID,
			bankCode:   account.BeneficiaryBankCode,
			accountNo:  account.BeneficiaryAccountNo,
			expected:   nil,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "account_inquiries")
			userRes, err := repo.GetByBankCodeAndAccountNo(ctx, tc.merchantId, tc.bankCode, tc.accountNo)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}
