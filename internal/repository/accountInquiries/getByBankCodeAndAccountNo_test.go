package accountInquiries

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/accountInquiries"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountInquiriesRepository_GetByBankCodeAndAccountNo(t *testing.T) {
	account := &accountInquiries.AccountInquiries{
		UUID:                   "uuid-uuid-uuid",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryAccountName: "testing",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing bank",
		Response:               "{}",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		bankCode  string
		accountNo string
		expected  *accountInquiries.AccountInquiries
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Account by Bank Code and Account No",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*accountInquiries.AccountInquiries)
					*userPtr = *account
				})
			},
			bankCode:  account.BeneficiaryBankCode,
			accountNo: account.BeneficiaryAccountNo,
			expected:  account,
			wantErr:   false,
		},
		{
			name: "ERROR: Account Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(sql.ErrNoRows)
			},
			bankCode: account.BeneficiaryBankCode,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(errors.New("database error"))

			},
			bankCode: account.BeneficiaryBankCode,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "account_inquiries")
			userRes, err := repo.GetByBankCodeAndAccountNo(ctx, tc.bankCode, tc.accountNo)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}
