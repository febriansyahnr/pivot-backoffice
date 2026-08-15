package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCardFundedPayoutDetail(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *cardFundedPayoutModel.GetPayoutDetailRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get detail with status history",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Mock GetContext for main data
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*cardFundedPayoutModel.GetPayoutDetailResponse"),
					mock.AnythingOfType("string"),
					"payout-123",
					"merchant-123",
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*cardFundedPayoutModel.GetPayoutDetailResponse)
					dest.UUID = "payout-123"
				}).Return(nil)

				// Mock SelectContext for status history
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]*statusHistoryModel.StatusHistory"),
					mock.AnythingOfType("string"),
					"payout-123",
					constant.DisbursementTypeCardFundedPayout,
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*statusHistoryModel.StatusHistory)
					*dest = append(*dest, &statusHistoryModel.StatusHistory{
						Status: "WAITING",
					})
				}).Return(nil)

				// Mock SelectContext for status history
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					"merchant-123",
					"payout-123",
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]string)
					*dest = append(*dest, uuid.NewString())
				}).Return(nil)
			},
			filter: &cardFundedPayoutModel.GetPayoutDetailRequest{
				PayoutID:   "payout-123",
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Payout not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*cardFundedPayoutModel.GetPayoutDetailResponse"),
					mock.AnythingOfType("string"),
					"not-found",
					"merchant-123",
				).Return(sql.ErrNoRows)
			},
			filter: &cardFundedPayoutModel.GetPayoutDetailRequest{
				PayoutID:   "not-found",
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error on main data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*cardFundedPayoutModel.GetPayoutDetailResponse"),
					mock.AnythingOfType("string"),
					"payout-123",
					"merchant-123",
				).Return(errors.New("db error"))
			},
			filter: &cardFundedPayoutModel.GetPayoutDetailRequest{
				PayoutID:   "payout-123",
				MerchantID: "merchant-123",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			result, err := repo.GetCardFundedPayoutDetail(ctx, tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
