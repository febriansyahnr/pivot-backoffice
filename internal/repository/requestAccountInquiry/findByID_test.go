package requestaccountinquiry

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFindByID(t *testing.T) {
	existedInquiry := &requestAccountInquiry.RequestAccountInquiryWithMaster{
		RequestAccountInquiries: requestAccountInquiry.RequestAccountInquiries{
			UUID:       uuid.NewString(),
			MerchantID: uuid.NewString(),
			AccountInquiryId: sql.NullString{
				String: uuid.NewString(),
				Valid:  true,
			},
			BeneficiaryBankCode: "008",
			BeneficiaryAccountNo: sql.NullString{
				String: "00809090909090909",
				Valid:  true,
			},
			CreatedAt: time.Now(),
		},
		MasterBeneficiaryAccountName: "John Wicks",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: FindByID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					existedPtr := args.Get(1).(*requestAccountInquiry.RequestAccountInquiryWithMaster)
					*existedPtr = *existedInquiry
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Inquiry Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)

			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: FindByID with valid metadata",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					existedPtr := args.Get(1).(*requestAccountInquiry.RequestAccountInquiryWithMaster)
					*existedPtr = *existedInquiry
					existedPtr.Metadata = types.NullJSONText{
						JSONText: []byte(`{"channelCode":"VA","additionalInfo":{"key":"value"}}`),
						Valid:    true,
					}
				})
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			_, err := repo.FindByID(ctx, uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
