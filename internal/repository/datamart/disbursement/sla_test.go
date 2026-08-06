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

func TestGetSLAMetrics_Integration(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	db := bqMock.NewIBigQueryService(t)

	repo := New(config.BigQueryConfig{PayoutSuccessMetricsTable: "test_table"}, db, logger)

	filter := disbursementModel.GetDisbursementInsightFilter{
		MerchantID:       "db7b3abe-f68d-40a7-b210-1ea84136bc03", // NOSONAR
		InsightStartDate: mustParseTimeSLA("2024-01-08T00:00:00Z"),
		InsightEndDate:   mustParseTimeSLA("2024-01-14T23:59:59Z"),
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  bool
		validation func(t *testing.T, result *disbursementModel.SLAMetrics)
	}{
		{
			name: "SUCCESS: Integration with real SLA data structure", // NOSONAR
			setupMock: func() {
				// Mock BigQuery response with realistic SLA data
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"transaction_date":  "2024-01-14",
							"metric_numerator":  2.1,  // Processing time in minutes
							"metric_denominator": nil, // SLA metrics don't use denominator
						},
						{
							"transaction_date":  "2024-01-13", 
							"metric_numerator":  2.3,
							"metric_denominator": nil,
						},
						{
							"transaction_date":  "2024-01-12",
							"metric_numerator":  1.9,
							"metric_denominator": nil,
						},
					},
					TotalRows: 3,
				}, nil)
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SLAMetrics) {
				assert.NotNil(t, result)
				assert.Len(t, result.DailyBreakdown, 3)
				// Verify average processing time calculation: (2.1+2.3+1.9)/3 = 2.1
				assert.InDelta(t, 2.1, result.AverageProcessingTimeMinutes, 0.01)
				// Verify daily breakdown is sorted by date (descending)
				assert.Equal(t, "2024-01-14", result.DailyBreakdown[0].Date)
				assert.InDelta(t, 2.1, result.DailyBreakdown[0].AverageProcessingTimeMinutes, 0.01)
				assert.Equal(t, "2024-01-13", result.DailyBreakdown[1].Date)
				assert.InDelta(t, 2.3, result.DailyBreakdown[1].AverageProcessingTimeMinutes, 0.01)
			},
		},
		{
			name: "SUCCESS: Handle zero processing times", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&bigquery.QueryResult{
					Rows: []map[string]any{
						{
							"transaction_date":  "2024-01-14",
							"metric_numerator":  0.0,
							"metric_denominator": nil,
						},
					},
					TotalRows: 1,
				}, nil)
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SLAMetrics) {
				assert.NotNil(t, result)
				assert.Len(t, result.DailyBreakdown, 1)
				assert.Equal(t, 0.0, result.AverageProcessingTimeMinutes)
				assert.Equal(t, "2024-01-14", result.DailyBreakdown[0].Date)
				assert.Equal(t, 0.0, result.DailyBreakdown[0].AverageProcessingTimeMinutes)
			},
		},
		{
			name: "SUCCESS: Handle empty result set", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&bigquery.QueryResult{
					Rows:      []map[string]any{},
					TotalRows: 0,
				}, nil)
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SLAMetrics) {
				assert.NotNil(t, result)
				assert.Empty(t, result.DailyBreakdown)
				assert.Equal(t, 0.0, result.AverageProcessingTimeMinutes)
			},
		},
		{
			name: "SUCCESS: Graceful handling when BigQuery service unavailable", // NOSONAR
			setupMock: func() {
				// Create repo with nil BigQuery service to simulate service unavailability
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SLAMetrics) {
				// Should get default metrics when BigQuery is unavailable
				nilRepo := New(config.BigQueryConfig{}, nil, logger)
				logger.On("Info", mock.Anything, "BigQuery service not configured, returning default SLA metrics").Once()
				
				result, err := nilRepo.GetSLAMetrics(context.Background(), filter)
				assert.NoError(t, err)
				assert.Equal(t, 0.0, result.AverageProcessingTimeMinutes)
				assert.Empty(t, result.DailyBreakdown)
			},
		},
		{
			name: "SUCCESS: Handle BigQuery execution error gracefully", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecuteQueryWithParams", mock.Anything, mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "failed to execute BigQuery SLA metrics query", mock.Anything,
				).Once().Return()
			},
			wantError: false,
			validation: func(t *testing.T, result *disbursementModel.SLAMetrics) {
				// Should return default metrics on BigQuery error
				assert.NotNil(t, result)
				assert.Equal(t, 0.0, result.AverageProcessingTimeMinutes)
				assert.Empty(t, result.DailyBreakdown)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name != "SUCCESS: Graceful handling when BigQuery service unavailable" {
				test.setupMock()
				result, err := repo.GetSLAMetrics(context.Background(), filter)
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
			if logger != nil {
				logger.AssertExpectations(t)
			}
		})
	}
}

// Helper function to parse time strings for SLA tests
func mustParseTimeSLA(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}