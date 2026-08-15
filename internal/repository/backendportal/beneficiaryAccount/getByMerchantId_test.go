package beneficiaryAccountRepository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/beneficiaryAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func TestGetByMerchantID(t *testing.T) {
	account := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryAccountName: "testing",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing bank",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		expected   *beneficiaryAccountModel.BeneficiaryAccount
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*beneficiaryAccountModel.BeneficiaryAccount)
					*userPtr = *account
				}).Once()
			},
			merchantId: account.MerchantID,
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
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(sql.ErrNoRows).Once()
			},
			merchantId: "non-existent-merchant-id",
			expected:   nil,
			wantErr:    false,
		},
		{
			name: "ERROR: error occurred Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			merchantId: "invalid-merchant-id",
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")
			userRes, err := repo.GetByMerchantID(ctx, tc.merchantId)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}
