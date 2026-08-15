package disbursementRepository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSumAmountByIDs(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		inputIDs  []string
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.SumAmountResponse"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr:  false,
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.SumAmountResponse"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr:  true,
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.SumAmountByIDs(ctx, tc.inputIDs)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetBeneficiaryTransactionLimit(t *testing.T) {
	startAt := time.Now().AddDate(0, -1, 0)
	endAt := time.Now()
	merchantId := uuid.NewString()

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		input     struct {
			merchantId string
			bankCode   string
			accountNo  string
			startAt    time.Time
			endAt      time.Time
		}
	}{
		{
			name: "SUCCESS: With Merchant ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.BeneficiaryPayoutLimitRuleLimit"),
					mock.AnythingOfType("string"),
					startAt, endAt, startAt, endAt,
					mock.Anything, // bankCode
					mock.Anything, // accountNo
					mock.Anything, // disbursementUpdatedAt
					merchantId, merchantId,
				).Return(nil).Run(func(args mock.Arguments) {
					arg := args[1].(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
					arg.Count = 5
					arg.Processed = 10000
				})
			},
			wantErr: false,
			input: struct {
				merchantId string
				bankCode   string
				accountNo  string
				startAt    time.Time
				endAt      time.Time
			}{
				merchantId: merchantId,
				bankCode:   "BCA",
				accountNo:  "1234567890",
				startAt:    startAt,
				endAt:      endAt,
			},
		},
		{
			name: "SUCCESS: With No Merchant ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.BeneficiaryPayoutLimitRuleLimit"),
					mock.AnythingOfType("string"),
					startAt, endAt, startAt, endAt,
					mock.Anything, // bankCode
					mock.Anything, // accountNo
					mock.Anything, // disbursementUpdatedAt
				).Return(nil).Run(func(args mock.Arguments) {
					arg := args[1].(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
					arg.Count = 3
					arg.Processed = 5000
				})
			},
			wantErr: false,
			input: struct {
				merchantId string
				bankCode   string
				accountNo  string
				startAt    time.Time
				endAt      time.Time
			}{
				bankCode:  "BCA",
				accountNo: "1234567890",
				startAt:   startAt,
				endAt:     endAt,
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.BeneficiaryPayoutLimitRuleLimit"),
					mock.AnythingOfType("string"),
					startAt, endAt, startAt, endAt,
					mock.Anything, // bankCode
					mock.Anything, // accountNo
					mock.Anything, // disbursementUpdatedAt
					merchantId, merchantId,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			input: struct {
				merchantId string
				bankCode   string
				accountNo  string
				startAt    time.Time
				endAt      time.Time
			}{
				merchantId: merchantId,
				bankCode:   "BCA",
				accountNo:  "1234567890",
				startAt:    startAt,
				endAt:      endAt,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetBeneficiaryTransactionLimit(
				ctx,
				tc.input.merchantId,
				tc.input.bankCode,
				tc.input.accountNo,
				tc.input.startAt,
				tc.input.endAt,
			)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.GreaterOrEqual(t, int64(result.Count), int64(0))
				assert.GreaterOrEqual(t, result.Processed, float64(0))
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
