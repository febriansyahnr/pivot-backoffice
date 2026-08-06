package paymentDatamart_test

import (
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart/payment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	bqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryPaymentSuccessRateComparison(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	db := bqMock.NewIBigQueryService(t)

	repo := New(config.BigQueryConfig{}, db, logger)

	request := paymentModel.QueryPaymentSuccessRateComparisonRequest{
		MerchantId: "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		PrevRange: paymentModel.DateRange{
			StartDate: "2025-11-17", // NOSONAR
			EndDate:   "2025-11-23", // NOSONAR
		},
		CurrentRange: paymentModel.DateRange{
			StartDate: "2025-11-24", // NOSONAR
			EndDate:   "2025-11-30", // NOSONAR
		},
	}
	params := map[string]any{
		"merchant_id":        "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		"prev_start_date":    "2025-11-17",                           // NOSONAR
		"prev_end_date":      "2025-11-23",                           // NOSONAR
		"current_start_date": "2025-11-24",                           // NOSONAR
		"current_end_date":   "2025-11-30",                           // NOSONAR
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *paymentModel.PaymentSuccessRateComparison
	}{
		{
			name: "ERROR:Execution query", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "failed to execute payment success rate query", mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:JSON binding", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previousSuccessRate": map[string]any{},
						},
					},
					TotalRows: 1,
				}, nil)
				logger.On(
					"Error", mock.Anything, "failed while binding the result value to the struct", mock.Anything,
				).Once().Return()
			},
			wantError: errors.New("JSON Binding: json: cannot unmarshal object into Go struct field PaymentSuccessRateComparison.previousSuccessRate of type json.Number"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previousSuccessRate": "99.00",  // NOSONAR
							"currentSuccessRate":  "100.00", // NOSONAR
							"differenceRate":      0,        // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &paymentModel.PaymentSuccessRateComparison{
				PreviousSuccessRate: "99.00",  // NOSONAR
				CurrentSuccessRate:  "100.00", // NOSONAR
				DifferenceRate:      "0",      // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.QueryPaymentSuccessRateComparison(t.Context(), request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
