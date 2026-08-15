package paymentRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigationSummary(t *testing.T) {
	merchantID := "merchant-uuid-123"
	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     paymentModel.GetInvestigationSummaryOption
		expected  *paymentModel.InvestigationSummaryResponse
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get investigation summary with date range",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.SummaryRow"),
					mock.AnythingOfType("string"),
					merchantID,
					startDate,
					endDate,
					startDate,
				).Return(nil).Run(func(args mock.Arguments) {
					row := args.Get(1).(*paymentModel.SummaryRow)
					row.TotalInProgress = 5000000
					row.TotalSuccess = 10000000
					row.TotalFailed = 2000000
					row.Currency = "IDR"
				})
			},
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			expected: &paymentModel.InvestigationSummaryResponse{
				OnInvestigation: paymentModel.InvestigationSummaryItem{
					TotalAmount: "5000000.00",
					Currency:    "IDR",
				},
				Success: paymentModel.InvestigationSummaryItem{
					TotalAmount: "10000000.00",
					Currency:    "IDR",
				},
				Failed: paymentModel.InvestigationSummaryItem{
					TotalAmount: "2000000.00",
					Currency:    "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No investigations found - zero amounts",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.SummaryRow"),
					mock.AnythingOfType("string"),
					merchantID,
					startDate,
					endDate,
					startDate,
				).Return(nil).Run(func(args mock.Arguments) {
					row := args.Get(1).(*paymentModel.SummaryRow)
					row.TotalInProgress = 0
					row.TotalSuccess = 0
					row.TotalFailed = 0
					row.Currency = "IDR"
				})
			},
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			expected: &paymentModel.InvestigationSummaryResponse{
				OnInvestigation: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0.00",
					Currency:    "IDR",
				},
				Success: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0.00",
					Currency:    "IDR",
				},
				Failed: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0.00",
					Currency:    "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Large amounts with proper formatting",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.SummaryRow"),
					mock.AnythingOfType("string"),
					merchantID,
					startDate,
					endDate,
					startDate,
				).Return(nil).Run(func(args mock.Arguments) {
					row := args.Get(1).(*paymentModel.SummaryRow)
					row.TotalInProgress = 999999999.99
					row.TotalSuccess = 1234567890.50
					row.TotalFailed = 5432100.75
					row.Currency = "IDR"
				})
			},
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			expected: &paymentModel.InvestigationSummaryResponse{
				OnInvestigation: paymentModel.InvestigationSummaryItem{
					TotalAmount: "999999999.99",
					Currency:    "IDR",
				},
				Success: paymentModel.InvestigationSummaryItem{
					TotalAmount: "1234567890.50",
					Currency:    "IDR",
				},
				Failed: paymentModel.InvestigationSummaryItem{
					TotalAmount: "5432100.75",
					Currency:    "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database connection error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.SummaryRow"),
					mock.AnythingOfType("string"),
					merchantID,
					startDate,
					endDate,
					startDate,
				).Return(errors.New("database connection failed"))
			},
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "ERROR: Query timeout",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.SummaryRow"),
					mock.AnythingOfType("string"),
					merchantID,
					startDate,
					endDate,
					startDate,
				).Return(errors.New("context deadline exceeded"))
			},
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mysqlMock := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := &PaymentRepository{
				db:     mysqlMock,
				logger: mockLogger,
			}

			// Execute
			result, err := repo.GetInvestigationSummary(context.Background(), tc.input)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.OnInvestigation.TotalAmount, result.OnInvestigation.TotalAmount)
				assert.Equal(t, tc.expected.OnInvestigation.Currency, result.OnInvestigation.Currency)
				assert.Equal(t, tc.expected.Success.TotalAmount, result.Success.TotalAmount)
				assert.Equal(t, tc.expected.Success.Currency, result.Success.Currency)
				assert.Equal(t, tc.expected.Failed.TotalAmount, result.Failed.TotalAmount)
				assert.Equal(t, tc.expected.Failed.Currency, result.Failed.Currency)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
