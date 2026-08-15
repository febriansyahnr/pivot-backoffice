package disbursementService

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) GetDisbursementInsight(ctx context.Context, filterRequest disbursementModel.GetDisbursementInsightFilter) (*disbursementModel.DisbursementInsightResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetDisbursementInsight")
	defer segment.End()

	var (
		defaultAmount = commonModel.Amount{
			Currency: "IDR",
			Value:    "0.00",
		}
		defaultSummaryTransaction = disbursementModel.SummaryTransaction{
			Count: 0,
			Sum:   defaultAmount,
		}

		waitingForApproval disbursementModel.SummaryTransaction
		delayed            disbursementModel.SummaryTransaction
		successTransaction disbursementModel.SummaryTransaction
		pendingTransaction disbursementModel.SummaryTransaction
		failedTransaction  disbursementModel.SummaryTransaction
		failureReasons     []disbursementModel.SummaryTransactionByReason
		successMetrics     *disbursementModel.SuccessRateMetrics
		slaMetrics         *disbursementModel.SLAMetrics

		response = &disbursementModel.DisbursementInsightResponse{
			WaitingForApproval: defaultSummaryTransaction,
			Delayed:            defaultSummaryTransaction,
			AllStatus: disbursementModel.AllStatusSummary{
				Success: defaultSummaryTransaction,
				Pending: defaultSummaryTransaction,
				Failed:  defaultSummaryTransaction,
			},
			FailureReasons: []disbursementModel.SummaryTransactionByReason{},
			SuccessMetrics: *s.getDefaultSuccessRateMetrics(),
			SLAMetrics:     *s.getDefaultSLAMetrics(),
		}
	)

	wg := sync.WaitGroup{}

	uuidOfMerchantID, err := uuid.Parse(filterRequest.MerchantID)
	if err != nil {
		return nil, err
	}

	// Calculate previous period dates for comparison
	var prevStartDate, prevEndDate time.Time
	if filterRequest.IncludePreviousPeriod {
		duration := filterRequest.InsightEndDate.Sub(filterRequest.InsightStartDate)
		prevStartDate = filterRequest.InsightStartDate.Add(-duration - time.Second)
		prevEndDate = filterRequest.InsightStartDate.Add(-time.Second)
	}

	// Find disbursement account by merchantID
	disbursementAccount, err := s.accountRepo.FindMerchantAccountByName(ctx, uuidOfMerchantID, constant.TypeDisbursement)
	if err != nil {
		return response, nil
	}
	if disbursementAccount == nil {
		err = constant.ErrMerchantBalanceNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return response, nil
	}

	dashboardFilter := disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID:       filterRequest.MerchantID,
		InsightStartDate: filterRequest.InsightStartDate,
		InsightEndDate:   filterRequest.InsightEndDate,
		IsXbPayout:       filterRequest.IsXbPayout,
	}

	// get WaitingForApproval
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryByDisbursementStatus(ctx, dashboardFilter, constant.DisbursementStatusWaiting)
		waitingForApproval = disbursementModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}

		response.WaitingForApproval = waitingForApproval
	}()

	// get Delayed (pending transactions with DELAYED reason)
	wg.Add(1)
	go func() {
		defer wg.Done()

		summaries, err := s.disbursementRepo.GetSummaryByReasonType(ctx, dashboardFilter, constant.StatusPending)
		if err != nil {
			s.logger.Error(ctx, "error when getting delayed disbursements", logger.Error(err))
			return
		}

		// Find DELAYED reason type and extract its count and sum
		for _, summary := range summaries {
			if summary.ReasonType == constant.ReasonTypePayoutDelayed {
				delayed = disbursementModel.SummaryTransaction{
					Count: summary.Count,
					Sum: commonModel.Amount{
						Currency: disbursementAccount.Currency,
						Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
					},
				}

				response.Delayed = delayed
				break
			}
		}
	}()

	// get Success Transaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummarySuccess(ctx, dashboardFilter)
		successTransaction = disbursementModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}

		response.AllStatus.Success = successTransaction
	}()

	// get Pending Transaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryInProgress(ctx, dashboardFilter)
		pendingTransaction = disbursementModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}

		response.AllStatus.Pending = pendingTransaction
	}()

	// get Failed Transaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryFailed(ctx, dashboardFilter)
		failedTransaction = disbursementModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}

		response.AllStatus.Failed = failedTransaction
	}()

	// get Failure Reasons
	wg.Add(1)
	go func() {
		defer wg.Done()

		summaries, err := s.disbursementRepo.GetSummaryByReasonType(ctx, dashboardFilter, constant.StatusFailed)
		if err != nil {
			s.logger.Error(ctx, "error when getting failed disbursements by reason type", logger.Error(err))
			return
		}

		for _, summary := range summaries {
			failureReasons = append(failureReasons, disbursementModel.SummaryTransactionByReason{
				ReasonType: summary.ReasonType,
				Count:      summary.Count,
				Sum: commonModel.Amount{
					Currency: disbursementAccount.Currency,
					Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
				},
			})
		}

		response.FailureReasons = failureReasons
	}()

	// get Success Rate Metrics from BigQuery
	// Keep this for backwards compatibility
	wg.Add(1)
	go func() {
		defer wg.Done()

		if s.disbursementMetricsRepo != nil {
			metrics, err := s.disbursementMetricsRepo.GetSuccessRateMetrics(ctx, filterRequest)
			if err != nil {
				s.logger.Error(ctx, "error when getting success rate metrics from BigQuery", logger.Error(err))
				successMetrics = s.getDefaultSuccessRateMetrics()
			} else {
				successMetrics = metrics
			}
		} else {
			successMetrics = s.getDefaultSuccessRateMetrics()
		}

		response.SuccessMetrics = *successMetrics
	}()

	// get SLA Metrics from BigQuery
	// Keep this for backwards compatibility
	wg.Add(1)
	go func() {
		defer wg.Done()

		if s.disbursementMetricsRepo != nil {
			metrics, err := s.disbursementMetricsRepo.GetSLAMetrics(ctx, filterRequest)
			if err != nil {
				s.logger.Error(ctx, "error when getting SLA metrics from BigQuery", logger.Error(err))
				slaMetrics = s.getDefaultSLAMetrics()
			} else {
				slaMetrics = metrics
			}
		} else {
			slaMetrics = s.getDefaultSLAMetrics()
		}

		response.SLAMetrics = *slaMetrics
	}()

	// get Success Rate Comparison (if enabled)
	if filterRequest.IncludePreviousPeriod && s.disbursementMetricsRepo != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			comparisonRequest := disbursementModel.QueryDisbursementSuccessRateComparisonRequest{
				MerchantId: filterRequest.MerchantID,
				PrevRange: disbursementModel.DateRange{
					StartDate: prevStartDate.Format(constant.DateFormat),
					EndDate:   prevEndDate.Format(constant.DateFormat),
				},
				CurrentRange: disbursementModel.DateRange{
					StartDate: filterRequest.InsightStartDate.Format(constant.DateFormat),
					EndDate:   filterRequest.InsightEndDate.Format(constant.DateFormat),
				},
			}

			successRateComparison, err := s.disbursementMetricsRepo.QueryDisbursementSuccessRateComparison(ctx, comparisonRequest)
			if err != nil {
				s.logger.Error(ctx, "error when getting success rate comparison from BigQuery", logger.Error(err))
			} else {
				response.SuccessRate = successRateComparison
			}
		}()

		// get SLA Comparison (if enabled)
		wg.Add(1)
		go func() {
			defer wg.Done()

			comparisonRequest := disbursementModel.QueryDisbursementSLAComparisonRequest{
				MerchantId: filterRequest.MerchantID,
				PrevRange: disbursementModel.DateRange{
					StartDate: prevStartDate.Format(constant.DateFormat),
					EndDate:   prevEndDate.Format(constant.DateFormat),
				},
				CurrentRange: disbursementModel.DateRange{
					StartDate: filterRequest.InsightStartDate.Format(constant.DateFormat),
					EndDate:   filterRequest.InsightEndDate.Format(constant.DateFormat),
				},
			}

			slaComparison, err := s.disbursementMetricsRepo.QueryDisbursementSLAComparison(ctx, comparisonRequest)
			if err != nil {
				s.logger.Error(ctx, "error when getting SLA comparison from BigQuery", logger.Error(err))
			} else {
				response.SLA = slaComparison
			}
		}()
	}

	wg.Wait()

	return response, nil
}

func (s *DisbursementService) getDefaultSuccessRateMetrics() *disbursementModel.SuccessRateMetrics {
	return &disbursementModel.SuccessRateMetrics{
		OverallSuccessRate: 0.0,
		AverageSuccessRate: 0.0,
		DailyBreakdown:     []disbursementModel.DailySuccessRateMetric{},
	}
}

func (s *DisbursementService) getDefaultSLAMetrics() *disbursementModel.SLAMetrics {
	return &disbursementModel.SLAMetrics{
		AverageProcessingTimeMinutes: 0.0,
		DailyBreakdown:               []disbursementModel.DailySLAMetric{},
	}
}
