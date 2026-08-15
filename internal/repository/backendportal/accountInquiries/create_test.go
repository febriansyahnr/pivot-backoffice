package accountInquiries

import (
	"context"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/accountInquiries"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func TestAccountInquiriesRepository_Create(t *testing.T) {
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
		account   *accountInquiries.AccountInquiries
		wantErr   bool
	}{
		{
			name:    "Valid Account",
			account: account,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
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
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
					).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"NamedExecContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType(constant.MockTypeAccountInquiriesReference),
					).
					Return(false, nil)
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
			err := repo.Create(context.Background(), tc.account)

			if (err != nil) != tc.wantErr {
				t.Errorf("AccountInquiries.Create() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
