package disbursementRepository

import (
	"context"
	"errors"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCardFundedPayoutList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *cardFundedPayoutModel.FilterGetPayoutList
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get list with data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]*cardFundedPayoutModel.GetPayoutListResponse"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]*cardFundedPayoutModel.GetPayoutListResponse)
					*dest = append(*dest, &cardFundedPayoutModel.GetPayoutListResponse{
						UUID: "payout-123",
					})
				}).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 1
				}).Return(nil)
			},
			filter: &cardFundedPayoutModel.FilterGetPayoutList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get empty list",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]*cardFundedPayoutModel.GetPayoutListResponse"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			filter: &cardFundedPayoutModel.FilterGetPayoutList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: GetContext (count) error is swallowed",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]*cardFundedPayoutModel.GetPayoutListResponse"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("db error count"))
			},
			filter: &cardFundedPayoutModel.FilterGetPayoutList{
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "ERROR: SelectContext (data) error is returned",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]*cardFundedPayoutModel.GetPayoutListResponse"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("db error data"))

				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &cardFundedPayoutModel.FilterGetPayoutList{
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
			result, err := repo.GetCardFundedPayoutList(ctx, tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
