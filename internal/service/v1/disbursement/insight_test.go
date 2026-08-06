package disbursementService

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDisbursementInsight(t *testing.T) {
	loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name             string
		filterRequest    disbursementModel.GetDisbursementInsightFilter
		setupMocks       func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository, metricsRepo *mocks.IDatamartDisbursementMetrics)
		expectedResponse *disbursementModel.DisbursementInsightResponse
		wantErr          bool
		expectedError    error
	}{
		{
			name: "SUCCESS: Get disbursement insight without comparison",
			filterRequest: disbursementModel.GetDisbursementInsightFilter{
				MerchantID:            uuid.New().String(),
				InsightStartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				InsightEndDate:        time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
				IsXbPayout:            false,
				IncludePreviousPeriod: false,
			},
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository, metricsRepo *mocks.IDatamartDisbursementMetrics) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(&account_model.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				// Mock all disbursement repository methods
				mockSummaryDTO := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 5,
					Sum:   100000.00,
				}

				disbursementRepo.On("GetSummaryByDisbursementStatus", mock.Anything, mock.Anything, constant.DisbursementStatusWaiting).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryByReasonType", mock.Anything, mock.Anything, constant.StatusPending).
					Return([]disbursementDashboardModel.SummaryTransactionByReasonType{
						{ReasonType: constant.ReasonTypePayoutDelayed, Count: 2, Sum: 50000.00},
					}, nil)
				disbursementRepo.On("GetSummarySuccess", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryInProgress", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryFailed", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryByReasonType", mock.Anything, mock.Anything, constant.StatusFailed).
					Return([]disbursementDashboardModel.SummaryTransactionByReasonType{
						{ReasonType: "TIMEOUT", Count: 1, Sum: 25000.00},
					}, nil)

				// Mock metrics repository methods
				successMetrics := &disbursementModel.SuccessRateMetrics{
					OverallSuccessRate: 95.5,
					AverageSuccessRate: 94.2,
					DailyBreakdown: []disbursementModel.DailySuccessRateMetric{
						{Date: "2024-01-01", SuccessfulCount: 10, TotalCount: 10, SuccessRatePercent: 100.0},
					},
				}
				slaMetrics := &disbursementModel.SLAMetrics{
					AverageProcessingTimeMinutes: 2.5,
					DailyBreakdown: []disbursementModel.DailySLAMetric{
						{Date: "2024-01-01", AverageProcessingTimeMinutes: 2.5},
					},
				}

				metricsRepo.On("GetSuccessRateMetrics", mock.Anything, mock.Anything).Return(successMetrics, nil)
				metricsRepo.On("GetSLAMetrics", mock.Anything, mock.Anything).Return(slaMetrics, nil)
			},
			expectedResponse: &disbursementModel.DisbursementInsightResponse{
				WaitingForApproval: disbursementModel.SummaryTransaction{
					Count: 5,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "100000.00"},
				},
				Delayed: disbursementModel.SummaryTransaction{
					Count: 2,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "50000.00"},
				},
				AllStatus: disbursementModel.AllStatusSummary{
					Success: disbursementModel.SummaryTransaction{
						Count: 5,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "100000.00"},
					},
					Pending: disbursementModel.SummaryTransaction{
						Count: 5,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "100000.00"},
					},
					Failed: disbursementModel.SummaryTransaction{
						Count: 5,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "100000.00"},
					},
				},
				FailureReasons: []disbursementModel.SummaryTransactionByReason{
					{
						ReasonType: "TIMEOUT",
						Count:      1,
						Sum:        commonModel.Amount{Currency: "IDR", Value: "25000.00"},
					},
				},
				SuccessMetrics: disbursementModel.SuccessRateMetrics{
					OverallSuccessRate: 95.5,
					AverageSuccessRate: 94.2,
					DailyBreakdown: []disbursementModel.DailySuccessRateMetric{
						{Date: "2024-01-01", SuccessfulCount: 10, TotalCount: 10, SuccessRatePercent: 100.0},
					},
				},
				SLAMetrics: disbursementModel.SLAMetrics{
					AverageProcessingTimeMinutes: 2.5,
					DailyBreakdown: []disbursementModel.DailySLAMetric{
						{Date: "2024-01-01", AverageProcessingTimeMinutes: 2.5},
					},
				},
				SuccessRate: nil, // No comparison data when IncludePreviousPeriod is false
				SLA:         nil, // No comparison data when IncludePreviousPeriod is false
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get disbursement insight with comparison",
			filterRequest: disbursementModel.GetDisbursementInsightFilter{
				MerchantID:            uuid.New().String(),
				InsightStartDate:      time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
				InsightEndDate:        time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC),
				IsXbPayout:            false,
				IncludePreviousPeriod: true,
			},
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository, metricsRepo *mocks.IDatamartDisbursementMetrics) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(&account_model.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				// Mock all disbursement repository methods
				mockSummaryDTO := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 8,
					Sum:   150000.00,
				}

				disbursementRepo.On("GetSummaryByDisbursementStatus", mock.Anything, mock.Anything, constant.DisbursementStatusWaiting).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryByReasonType", mock.Anything, mock.Anything, constant.StatusPending).
					Return([]disbursementDashboardModel.SummaryTransactionByReasonType{
						{ReasonType: constant.ReasonTypePayoutDelayed, Count: 3, Sum: 75000.00},
					}, nil)
				disbursementRepo.On("GetSummarySuccess", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryInProgress", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryFailed", mock.Anything, mock.Anything).
					Return(mockSummaryDTO, nil)
				disbursementRepo.On("GetSummaryByReasonType", mock.Anything, mock.Anything, constant.StatusFailed).
					Return([]disbursementDashboardModel.SummaryTransactionByReasonType{
						{ReasonType: "TIMEOUT", Count: 2, Sum: 40000.00},
					}, nil)

				// Mock metrics repository methods
				successMetrics := &disbursementModel.SuccessRateMetrics{
					OverallSuccessRate: 96.8,
					AverageSuccessRate: 95.5,
					DailyBreakdown: []disbursementModel.DailySuccessRateMetric{
						{Date: "2024-01-08", SuccessfulCount: 15, TotalCount: 15, SuccessRatePercent: 100.0},
					},
				}
				slaMetrics := &disbursementModel.SLAMetrics{
					AverageProcessingTimeMinutes: 2.1,
					DailyBreakdown: []disbursementModel.DailySLAMetric{
						{Date: "2024-01-08", AverageProcessingTimeMinutes: 2.1},
					},
				}

				// Mock comparison methods
				successRateComparison := &disbursementModel.ComparisonData{
					Previous:   json.Number("94.2"),
					Current:    json.Number("96.8"),
					Difference: json.Number("2.6"),
				}
				slaComparison := &disbursementModel.ComparisonData{
					Previous:   json.Number("2.5"),
					Current:    json.Number("2.1"),
					Difference: json.Number("-0.4"),
				}

				metricsRepo.On("GetSuccessRateMetrics", mock.Anything, mock.Anything).Return(successMetrics, nil)
				metricsRepo.On("GetSLAMetrics", mock.Anything, mock.Anything).Return(slaMetrics, nil)
				metricsRepo.On("QueryDisbursementSuccessRateComparison", mock.Anything, mock.Anything).Return(successRateComparison, nil)
				metricsRepo.On("QueryDisbursementSLAComparison", mock.Anything, mock.Anything).Return(slaComparison, nil)
			},
			expectedResponse: &disbursementModel.DisbursementInsightResponse{
				WaitingForApproval: disbursementModel.SummaryTransaction{
					Count: 8,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "150000.00"},
				},
				Delayed: disbursementModel.SummaryTransaction{
					Count: 3,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "75000.00"},
				},
				AllStatus: disbursementModel.AllStatusSummary{
					Success: disbursementModel.SummaryTransaction{
						Count: 8,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "150000.00"},
					},
					Pending: disbursementModel.SummaryTransaction{
						Count: 8,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "150000.00"},
					},
					Failed: disbursementModel.SummaryTransaction{
						Count: 8,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "150000.00"},
					},
				},
				FailureReasons: []disbursementModel.SummaryTransactionByReason{
					{
						ReasonType: "TIMEOUT",
						Count:      2,
						Sum:        commonModel.Amount{Currency: "IDR", Value: "40000.00"},
					},
				},
				SuccessMetrics: disbursementModel.SuccessRateMetrics{
					OverallSuccessRate: 96.8,
					AverageSuccessRate: 95.5,
					DailyBreakdown: []disbursementModel.DailySuccessRateMetric{
						{Date: "2024-01-08", SuccessfulCount: 15, TotalCount: 15, SuccessRatePercent: 100.0},
					},
				},
				SLAMetrics: disbursementModel.SLAMetrics{
					AverageProcessingTimeMinutes: 2.1,
					DailyBreakdown: []disbursementModel.DailySLAMetric{
						{Date: "2024-01-08", AverageProcessingTimeMinutes: 2.1},
					},
				},
				SuccessRate: &disbursementModel.ComparisonData{
					Previous:   json.Number("94.2"),
					Current:    json.Number("96.8"),
					Difference: json.Number("2.6"),
				},
				SLA: &disbursementModel.ComparisonData{
					Previous:   json.Number("2.5"),
					Current:    json.Number("2.1"),
					Difference: json.Number("-0.4"),
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Invalid merchant ID",
			filterRequest: disbursementModel.GetDisbursementInsightFilter{
				MerchantID:            "invalid-uuid",
				InsightStartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				InsightEndDate:        time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
				IsXbPayout:            false,
				IncludePreviousPeriod: false,
			},
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository, metricsRepo *mocks.IDatamartDisbursementMetrics) {
				// No mocks needed for this error case
			},
			expectedResponse: nil,
			wantErr:          true,
		},
		{
			name: "ERROR: Account not found",
			filterRequest: disbursementModel.GetDisbursementInsightFilter{
				MerchantID:            uuid.New().String(),
				InsightStartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				InsightEndDate:        time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
				IsXbPayout:            false,
				IncludePreviousPeriod: false,
			},
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository, metricsRepo *mocks.IDatamartDisbursementMetrics) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(nil, nil) // Account not found
			},
			expectedResponse: &disbursementModel.DisbursementInsightResponse{
				WaitingForApproval: disbursementModel.SummaryTransaction{
					Count: 0,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
				},
				Delayed: disbursementModel.SummaryTransaction{
					Count: 0,
					Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
				},
				AllStatus: disbursementModel.AllStatusSummary{
					Success: disbursementModel.SummaryTransaction{
						Count: 0,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
					},
					Pending: disbursementModel.SummaryTransaction{
						Count: 0,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
					},
					Failed: disbursementModel.SummaryTransaction{
						Count: 0,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
					},
				},
				FailureReasons: []disbursementModel.SummaryTransactionByReason{},
				SuccessMetrics: disbursementModel.SuccessRateMetrics{
					OverallSuccessRate: 0.0,
					AverageSuccessRate: 0.0,
					DailyBreakdown:     []disbursementModel.DailySuccessRateMetric{},
				},
				SLAMetrics: disbursementModel.SLAMetrics{
					AverageProcessingTimeMinutes: 0.0,
					DailyBreakdown:               []disbursementModel.DailySLAMetric{},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepoMock := mocks.NewIDisbursementRepository(t)
			accountRepoMock := mocks.NewIAccountRepository(t)
			metricsRepoMock := mocks.NewIDatamartDisbursementMetrics(t)

			tc.setupMocks(disbursementRepoMock, accountRepoMock, metricsRepoMock)

			service := &DisbursementService{
				logger:                  loggerMock,
				disbursementRepo:        disbursementRepoMock,
				accountRepo:             accountRepoMock,
				disbursementMetricsRepo: metricsRepoMock,
			}

			response, err := service.GetDisbursementInsight(context.Background(), tc.filterRequest)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tc.expectedResponse.WaitingForApproval, response.WaitingForApproval)
				assert.Equal(t, tc.expectedResponse.Delayed, response.Delayed)
				assert.Equal(t, tc.expectedResponse.AllStatus, response.AllStatus)
				assert.Equal(t, tc.expectedResponse.FailureReasons, response.FailureReasons)
				assert.Equal(t, tc.expectedResponse.SuccessMetrics, response.SuccessMetrics)
				assert.Equal(t, tc.expectedResponse.SLAMetrics, response.SLAMetrics)
				assert.Equal(t, tc.expectedResponse.SuccessRate, response.SuccessRate)
				assert.Equal(t, tc.expectedResponse.SLA, response.SLA)
			}
		})
	}
}

func TestGetDisbursementInsight_PeriodCalculation(t *testing.T) {
	loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	t.Run("Period calculation logic", func(t *testing.T) {
		// Test that the previous period is calculated correctly
		filterRequest := disbursementModel.GetDisbursementInsightFilter{
			MerchantID:            uuid.New().String(),
			InsightStartDate:      time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),     // Week start
			InsightEndDate:        time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC), // Week end
			IsXbPayout:            false,
			IncludePreviousPeriod: true,
		}

		// Expected previous period: 2024-01-01 00:00:00 to 2024-01-07 23:59:58
		expectedPrevStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedPrevEnd := time.Date(2024, 1, 7, 23, 59, 58, 0, time.UTC)

		disbursementRepoMock := mocks.NewIDisbursementRepository(t)
		accountRepoMock := mocks.NewIAccountRepository(t)
		metricsRepoMock := mocks.NewIDatamartDisbursementMetrics(t)

		// Mock account lookup
		accountRepoMock.On(
			"FindMerchantAccountByName",
			mock.AnythingOfType(constant.MockTypeValueContextReference),
			constant.UuidMockType(),
			constant.StringMockType(),
		).Return(&account_model.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

		// Mock all required repository calls with minimal expectations
		disbursementRepoMock.On("GetSummaryByDisbursementStatus", mock.Anything, mock.Anything, mock.Anything).
			Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
		disbursementRepoMock.On("GetSummaryByReasonType", mock.Anything, mock.Anything, mock.Anything).
			Return([]disbursementDashboardModel.SummaryTransactionByReasonType{}, nil)
		disbursementRepoMock.On("GetSummarySuccess", mock.Anything, mock.Anything).
			Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
		disbursementRepoMock.On("GetSummaryInProgress", mock.Anything, mock.Anything).
			Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
		disbursementRepoMock.On("GetSummaryFailed", mock.Anything, mock.Anything).
			Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)

		metricsRepoMock.On("GetSuccessRateMetrics", mock.Anything, mock.Anything).
			Return(&disbursementModel.SuccessRateMetrics{}, nil)
		metricsRepoMock.On("GetSLAMetrics", mock.Anything, mock.Anything).
			Return(&disbursementModel.SLAMetrics{}, nil)

		// Capture the comparison request to verify period calculation
		metricsRepoMock.On("QueryDisbursementSuccessRateComparison", mock.Anything, mock.MatchedBy(
			func(req disbursementModel.QueryDisbursementSuccessRateComparisonRequest) bool {
				// Verify the previous period dates are calculated correctly
				expectedPrevStartStr := expectedPrevStart.Format("2006-01-02")
				expectedPrevEndStr := expectedPrevEnd.Format("2006-01-02")
				return req.PrevRange.StartDate == expectedPrevStartStr && req.PrevRange.EndDate == expectedPrevEndStr
			},
		)).Return(&disbursementModel.ComparisonData{}, nil)

		metricsRepoMock.On("QueryDisbursementSLAComparison", mock.Anything, mock.MatchedBy(
			func(req disbursementModel.QueryDisbursementSLAComparisonRequest) bool {
				// Verify the previous period dates are calculated correctly
				expectedPrevStartStr := expectedPrevStart.Format("2006-01-02")
				expectedPrevEndStr := expectedPrevEnd.Format("2006-01-02")
				return req.PrevRange.StartDate == expectedPrevStartStr && req.PrevRange.EndDate == expectedPrevEndStr
			},
		)).Return(&disbursementModel.ComparisonData{}, nil)

		service := &DisbursementService{
			logger:                  loggerMock,
			disbursementRepo:        disbursementRepoMock,
			accountRepo:             accountRepoMock,
			disbursementMetricsRepo: metricsRepoMock,
		}

		_, err := service.GetDisbursementInsight(context.Background(), filterRequest)

		assert.NoError(t, err)
		// Assertions are handled by the mock expectations
	})
}

