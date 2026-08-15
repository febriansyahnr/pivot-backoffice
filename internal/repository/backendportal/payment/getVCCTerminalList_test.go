package paymentRepository_test

import (
	"context"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVCCTerminalList(t *testing.T) {
	chargeStartDate := time.Now().UTC().Add(-24 * time.Hour)
	chargeEndDate := time.Now().UTC()

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *paymentModel.GetVCCTerminalListFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get VCC Terminal List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with sortBy createdAt",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				SortBy:          "createdAt",
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with sortBy chargeDate",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				SortBy:          "chargeDate",
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with pagination",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				Page:            2,
				PerPage:         20,
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with Page < 1 (defaults to 1)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				Page:            0,
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with PerPage < 1 (defaults to 10)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				PerPage:         0,
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with Sort provided (sets to ASC)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				Sort:            "DESC",
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with empty Sort (default)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				Sort:            "",
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get VCC Terminal List with data to trigger amount mapping",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					dataPtr := args.Get(1).(*[]*paymentModel.VccTerminalItem)
					*dataPtr = []*paymentModel.VccTerminalItem{
						{
							ChargeID:       "test-charge-1",
							ChargeAmount:   decimal.NewFromInt(100000),
							ChargeCurrency: "IDR",
							ChargeDate:     time.Now(),
							Status:         constant.UnifiedPaymentSessionStatusPaid,
							BulkID:         "bulk-1",
							TravelAgent:    "Travel Agent 1",
						},
						{
							ChargeID:       "test-charge-2",
							ChargeAmount:   decimal.NewFromInt(200000),
							ChargeCurrency: "IDR",
							ChargeDate:     time.Now(),
							Status:         constant.UnifiedPaymentSessionStatusPaid,
							BulkID:         "bulk-2",
							TravelAgent:    "Travel Agent 2",
						},
					}
				})
			},
			filter: &paymentModel.GetVCCTerminalListFilterRequest{
				ChargeStartDate: chargeStartDate,
				ChargeEndDate:   chargeEndDate,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetVCCTerminalList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
