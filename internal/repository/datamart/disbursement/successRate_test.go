package disbursementDatamart_test

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart/disbursement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	bqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryDisbursementSuccessRateComparison(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	db := bqMock.NewIBigQueryService(t)

	repo := New(config.BigQueryConfig{}, db, logger)

	request := disbursementModel.QueryDisbursementSuccessRateComparisonRequest{
		MerchantId: "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		PrevRange: disbursementModel.DateRange{
			StartDate: "2024-01-01", // NOSONAR
			EndDate:   "2024-01-07", // NOSONAR
		},
		CurrentRange: disbursementModel.DateRange{
			StartDate: "2024-01-08", // NOSONAR
			EndDate:   "2024-01-14", // NOSONAR
		},
	}
	params := map[string]any{
		"merchant_id":        "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		"prev_start_date":    "2024-01-01",                           // NOSONAR
		"prev_end_date":      "2024-01-07",                           // NOSONAR
		"current_start_date": "2024-01-08",                           // NOSONAR
		"current_end_date":   "2024-01-14",                           // NOSONAR
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *disbursementModel.ComparisonData
	}{
		{
			name: "ERROR: BigQuery execution failure", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "failed to execute disbursement success rate comparison query", mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: JSON binding failure with invalid data", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   map[string]any{}, // Invalid type causes binding error
							"current":    "96.8",
							"difference": "2.6",
						},
					},
					TotalRows: 1,
				}, nil)
				logger.On(
					"Error", mock.Anything, "failed while binding the result value to the struct", mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: Valid comparison data with positive difference", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "94.2", // NOSONAR
							"current":    "96.8", // NOSONAR
							"difference": "2.6",  // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "94.2", // NOSONAR
				Current:    "96.8", // NOSONAR
				Difference: "2.6",  // NOSONAR
			},
		},
		{
			name: "SUCCESS: Valid comparison data with negative difference", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "97.5", // NOSONAR
							"current":    "95.1", // NOSONAR
							"difference": "-2.4", // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "97.5", // NOSONAR
				Current:    "95.1", // NOSONAR
				Difference: "-2.4", // NOSONAR
			},
		},
		{
			name: "SUCCESS: No data returned (empty result set)", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows:      []map[string]any{},
					TotalRows: 0,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "0",
				Current:    "0",
				Difference: "0",
			},
		},
		{
			name: "SUCCESS: Zero values with proper calculation", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "0.0", // NOSONAR
							"current":    "0.0", // NOSONAR
							"difference": "0.0", // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "0.0", // NOSONAR
				Current:    "0.0", // NOSONAR
				Difference: "0.0", // NOSONAR
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.QueryDisbursementSuccessRateComparison(context.Background(), request)
			if test.wantError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			}

			db.AssertExpectations(t)
			if test.name == "ERROR: BigQuery execution failure" || test.name == "ERROR: JSON binding failure with invalid data" {
				logger.AssertExpectations(t)
			}
		})
	}
}

func TestQueryDisbursementSLAComparison(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	db := bqMock.NewIBigQueryService(t)

	repo := New(config.BigQueryConfig{}, db, logger)

	request := disbursementModel.QueryDisbursementSLAComparisonRequest{
		MerchantId: "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		PrevRange: disbursementModel.DateRange{
			StartDate: "2024-01-01", // NOSONAR
			EndDate:   "2024-01-07", // NOSONAR
		},
		CurrentRange: disbursementModel.DateRange{
			StartDate: "2024-01-08", // NOSONAR
			EndDate:   "2024-01-14", // NOSONAR
		},
	}
	params := map[string]any{
		"merchant_id":        "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		"prev_start_date":    "2024-01-01",                           // NOSONAR
		"prev_end_date":      "2024-01-07",                           // NOSONAR
		"current_start_date": "2024-01-08",                           // NOSONAR
		"current_end_date":   "2024-01-14",                           // NOSONAR
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *disbursementModel.ComparisonData
	}{
		{
			name: "ERROR: BigQuery execution failure", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "failed to execute disbursement SLA comparison query", mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: JSON binding failure with invalid data", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   map[string]any{}, // Invalid type causes binding error
							"current":    "2.1",
							"difference": "-0.4",
						},
					},
					TotalRows: 1,
				}, nil)
				logger.On(
					"Error", mock.Anything, "failed while binding the result value to the struct", mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: Valid SLA comparison with processing time improvement", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "2.5",  // NOSONAR
							"current":    "2.1",  // NOSONAR
							"difference": "-0.4", // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "2.5",  // NOSONAR
				Current:    "2.1",  // NOSONAR
				Difference: "-0.4", // NOSONAR
			},
		},
		{
			name: "SUCCESS: Valid SLA comparison with processing time degradation", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "1.8", // NOSONAR
							"current":    "2.3", // NOSONAR
							"difference": "0.5", // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "1.8", // NOSONAR
				Current:    "2.3", // NOSONAR
				Difference: "0.5", // NOSONAR
			},
		},
		{
			name: "SUCCESS: No data returned (empty result set)", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows:      []map[string]any{},
					TotalRows: 0,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "0",
				Current:    "0",
				Difference: "0",
			},
		},
		{
			name: "SUCCESS: Zero processing times", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.Anything, params,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"previous":   "0.0", // NOSONAR
							"current":    "0.0", // NOSONAR
							"difference": "0.0", // NOSONAR
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantResult: &disbursementModel.ComparisonData{
				Previous:   "0.0", // NOSONAR
				Current:    "0.0", // NOSONAR
				Difference: "0.0", // NOSONAR
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.QueryDisbursementSLAComparison(context.Background(), request)
			if test.wantError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			}

			db.AssertExpectations(t)
			if test.name == "ERROR: BigQuery execution failure" || test.name == "ERROR: JSON binding failure with invalid data" {
				logger.AssertExpectations(t)
			}
		})
	}
}

func TestGetSuccessRateMetrics_Integration(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	db := bqMock.NewIBigQueryService(t)

	repo := New(config.BigQueryConfig{PayoutSuccessMetricsTable: "test_table"}, db, logger)

	filter := disbursementModel.GetDisbursementInsightFilter{
		MerchantID:       "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		InsightStartDate: mustParseTime("2024-01-08T00:00:00Z"),
		InsightEndDate:   mustParseTime("2024-01-14T23:59:59Z"),
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  bool
		validation func(t *testing.T, result *disbursementModel.SuccessRateMetrics)
	}{
		{
			name: "SUCCESS: Integration with real BigQuery data structure", // NOSONAR
			setupMock: func() {
				// Mock BigQuery response with realistic data
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"transaction_date":   "2024-01-14",
							"metric_numerator":   int64(18),
							"metric_denominator": int64(20),
						},
						{
							"transaction_date":   "2024-01-13",
							"metric_numerator":   int64(22),
							"metric_denominator": int64(25),
						},
						{
							"transaction_date":   "2024-01-12",
							"metric_numerator":   int64(30),
							"metric_denominator": int64(30),
						},
					},
					TotalRows: 3,
				}, nil)
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SuccessRateMetrics) {
				assert.NotNil(t, result)
				assert.Len(t, result.DailyBreakdown, 3)
				// Verify overall success rate calculation: (18+22+30)/(20+25+30) = 70/75 = 93.33%
				assert.InDelta(t, 93.33, result.OverallSuccessRate, 0.01)
				// Verify daily breakdown is sorted by date (descending)
				assert.Equal(t, "2024-01-14", result.DailyBreakdown[0].Date)
				assert.Equal(t, int64(18), result.DailyBreakdown[0].SuccessfulCount)
				assert.Equal(t, int64(20), result.DailyBreakdown[0].TotalCount)
				assert.InDelta(t, 90.0, result.DailyBreakdown[0].SuccessRatePercent, 0.01)
			},
		},
		{
			name: "SUCCESS: Graceful handling when BigQuery service unavailable", // NOSONAR
			setupMock: func() {
				// Create repo with nil BigQuery service to simulate service unavailability
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SuccessRateMetrics) {
				// Should get default metrics when BigQuery is unavailable
				nilRepo := New(config.BigQueryConfig{}, nil, logger)
				logger.On("Info", mock.Anything, "BigQuery service not configured, returning default success rate metrics").Once()

				result, err := nilRepo.GetSuccessRateMetrics(context.Background(), filter)
				assert.NoError(t, err)
				assert.Equal(t, 0.0, result.OverallSuccessRate)
				assert.Equal(t, 0.0, result.AverageSuccessRate)
				assert.Empty(t, result.DailyBreakdown)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name != "SUCCESS: Graceful handling when BigQuery service unavailable" {
				test.setupMock()
				result, err := repo.GetSuccessRateMetrics(context.Background(), filter)
				if test.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					test.validation(t, result)
				}
				if db != nil {
					db.AssertExpectations(t)
				}
			} else {
				test.validation(t, nil)
			}
		})
	}
}

// Helper function to parse time strings for tests
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}
