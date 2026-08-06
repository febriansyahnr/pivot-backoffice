package requestaccountinquiry

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindLatestByNumberAccount(t *testing.T) {
	accountNumber := "1234567890"
	merchantID := uuid.NewString()
	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(db *mysqlMocks.IMySqlExt)
	}{
		{
			desc:    "error when FindLatestByNumberAccount",
			wantErr: true,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiries"),
					mock.AnythingOfType("string"),
					merchantID,
					accountNumber,
				).Return(assert.AnError)
			},
		},
		{
			desc:    "success when FindLatestByNumberAccount",
			wantErr: false,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiries"),
					mock.AnythingOfType("string"),
					merchantID,
					accountNumber,
				).Return(nil).Run(func(args mock.Arguments) {
					requestAccountInquiries := args.Get(1).(*requestAccountInquiry.RequestAccountInquiries)
					*requestAccountInquiries = requestAccountInquiry.RequestAccountInquiries{
						UUID:                uuid.NewString(),
						MerchantID:          merchantID,
						AccountInquiryId:    sql.NullString{String: uuid.NewString(), Valid: true},
						BeneficiaryBankCode: "008",
						BeneficiaryAccountNo: sql.NullString{
							String: "00809090909090909",
							Valid:  true,
						},
					}
				})
			},
		},
		{
			desc:    "success when FindLatestByNumberAccount - no rows found",
			wantErr: false,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiries"),
					mock.AnythingOfType("string"),
					merchantID,
					accountNumber,
				).Return(sql.ErrNoRows)
			},
		},
		{
			desc:    "success when FindLatestByNumberAccount with valid metadata",
			wantErr: false,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiries"),
					mock.AnythingOfType("string"),
					merchantID,
					accountNumber,
				).Return(nil).Run(func(args mock.Arguments) {
					requestAccountInquiries := args.Get(1).(*requestAccountInquiry.RequestAccountInquiries)
					*requestAccountInquiries = requestAccountInquiry.RequestAccountInquiries{
						UUID:                uuid.NewString(),
						MerchantID:          merchantID,
						AccountInquiryId:    sql.NullString{String: uuid.NewString(), Valid: true},
						BeneficiaryBankCode: "008",
						BeneficiaryAccountNo: sql.NullString{
							String: "00809090909090909",
							Valid:  true,
						},
						Metadata: types.NullJSONText{
							JSONText: []byte(`{"channelCode":"VA","additionalInfo":{"key":"value"}}`),
							Valid:    true,
						},
					}
				})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			logger, _ := pdkLogger.NewZapLogger(pdkLogger.Config{})
			db := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(db)

			repo := New(db, logger)

			_, err := repo.FindLatestByNumberAccount(context.Background(), accountNumber, merchantID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
