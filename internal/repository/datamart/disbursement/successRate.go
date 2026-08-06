package disbursementDatamart

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const MetricNameSuccessRate = "Success Rate"

func (r *repository) getTableName() string {
	if r.config.PayoutSuccessMetricsTable != "" {
		return r.config.PayoutSuccessMetricsTable
	}
	return "data_mart_layer.rpt__payout_success_metrics__daily"
}

func (r *repository) getMetricsFromBigQuery(
	ctx context.Context,
	filter disbursementModel.GetDisbursementInsightFilter,
	metricName string,
) (*bigquery.QueryResult, error) {
	query := fmt.Sprintf(`
		SELECT 
			transaction_date,
			metric_numerator,
			metric_denominator
		FROM %s 
		WHERE 
			merchant_id = @merchant_id
			AND metric_name = @metric_name
			AND transaction_date >= @start_date 
			AND transaction_date <= @end_date
		ORDER BY transaction_date DESC
	`, r.getTableName())

	params := map[string]any{
		"merchant_id": filter.MerchantID,
		"metric_name": metricName,
		"start_date":  filter.InsightStartDate.Format("2006-01-02"),
		"end_date":    filter.InsightEndDate.Format("2006-01-02"),
	}

	return r.db.ExecuteQueryWithParams(ctx, query, params)
}

func (r *repository) GetSuccessRateMetrics(
	ctx context.Context,
	filter disbursementModel.GetDisbursementInsightFilter,
) (*disbursementModel.SuccessRateMetrics, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/datamart/disbursement/GetSuccessRateMetrics")
	defer segment.End()

	if r.db == nil {
		r.logger.Info(ctx, "BigQuery service not configured, returning default success rate metrics")
		return r.getDefaultSuccessRateMetrics(), nil
	}

	result, err := r.getMetricsFromBigQuery(ctx, filter, MetricNameSuccessRate)
	if err != nil {
		r.logger.Error(ctx, "failed to execute BigQuery success rate query", logger.Error(err))
		return r.getDefaultSuccessRateMetrics(), nil
	}

	metrics := r.getDefaultSuccessRateMetrics()

	if len(result.Rows) == 0 {
		return metrics, nil
	}

	var dailyBreakdown []disbursementModel.DailySuccessRateMetric
	var totalSuccessful, totalCount int64
	var successRateSum float64

	for _, row := range result.Rows {
		date := r.getStringFromInterface(row["transaction_date"])
		numerator := r.getInt64FromInterface(row["metric_numerator"])
		denominator := r.getInt64FromInterface(row["metric_denominator"])

		var successRate float64
		if denominator > 0 {
			successRate = (float64(numerator) / float64(denominator)) * 100.0
		}

		dailyMetric := disbursementModel.DailySuccessRateMetric{
			Date:               date,
			SuccessfulCount:    numerator,
			TotalCount:         denominator,
			SuccessRatePercent: successRate,
		}

		dailyBreakdown = append(dailyBreakdown, dailyMetric)

		totalSuccessful += numerator
		totalCount += denominator
		successRateSum += successRate
	}

	if totalCount > 0 {
		metrics.OverallSuccessRate = (float64(totalSuccessful) / float64(totalCount)) * 100.0
	}
	if len(dailyBreakdown) > 0 {
		metrics.AverageSuccessRate = successRateSum / float64(len(dailyBreakdown))
	}

	sort.Slice(dailyBreakdown, func(i, j int) bool {
		return dailyBreakdown[i].Date > dailyBreakdown[j].Date
	})

	metrics.DailyBreakdown = dailyBreakdown
	return metrics, nil
}

func (r *repository) getDefaultSuccessRateMetrics() *disbursementModel.SuccessRateMetrics {
	return &disbursementModel.SuccessRateMetrics{
		OverallSuccessRate: 0.0,
		AverageSuccessRate: 0.0,
		DailyBreakdown:     []disbursementModel.DailySuccessRateMetric{},
	}
}

func (r *repository) getStringFromInterface(val any) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case time.Time:
		return v.Format("2006-01-02")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (r *repository) getInt64FromInterface(val any) int64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func (r *repository) QueryDisbursementSuccessRateComparison(ctx context.Context, request disbursementModel.QueryDisbursementSuccessRateComparisonRequest) (*disbursementModel.ComparisonData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/datamart/disbursement/QueryDisbursementSuccessRateComparison")
	defer segment.End()

	if r.db == nil {
		r.logger.Info(ctx, "BigQuery service not configured, returning default comparison data")
		return &disbursementModel.ComparisonData{
			Previous:   "0",
			Current:    "0",
			Difference: "0",
		}, nil
	}

	params := map[string]any{
		"merchant_id":        request.MerchantId,
		"prev_start_date":    request.PrevRange.StartDate,
		"prev_end_date":      request.PrevRange.EndDate,
		"current_start_date": request.CurrentRange.StartDate,
		"current_end_date":   request.CurrentRange.EndDate,
	}

	rawQuery := fmt.Sprintf(`WITH disbursement_grouping_by_range AS (
		SELECT
			COALESCE(SUM(IF(transaction_date BETWEEN @prev_start_date AND @prev_end_date, metric_numerator, 0)), 0) AS prev_success_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @prev_start_date AND @prev_end_date, metric_denominator, 0)), 0) AS prev_total_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @current_start_date AND @current_end_date, metric_numerator, 0)), 0) AS curr_success_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @current_start_date AND @current_end_date, metric_denominator, 0)), 0) AS curr_total_count
		FROM 
			%s
		WHERE
			merchant_id = @merchant_id
			AND metric_name = 'Success Rate'
			AND transaction_date BETWEEN @prev_start_date AND @current_end_date
	), disbursement_success_rate AS (
		SELECT 
			ROUND(SAFE_DIVIDE(prev_success_count, prev_total_count) * 100, 2) AS previous_success_rate,
			ROUND(SAFE_DIVIDE(curr_success_count, curr_total_count) * 100, 2) AS current_success_rate
		FROM disbursement_grouping_by_range
	) SELECT 
		COALESCE(previous_success_rate, 0) AS `+"`previous`"+`, 
		COALESCE(current_success_rate, 0) AS `+"`current`"+`, 
		ROUND(COALESCE(current_success_rate, 0) - COALESCE(previous_success_rate, 0), 2) AS `+"`difference`"+`
	FROM disbursement_success_rate;`, r.getTableName())

	result, err := r.db.ExecuteQueryWithParams(ctx, rawQuery, params)
	if err != nil {
		r.logger.Error(ctx, "failed to execute disbursement success rate comparison query", logger.Error(err))
		return nil, err
	}

	comparisonData := &disbursementModel.ComparisonData{
		Previous:   "0",
		Current:    "0",
		Difference: "0",
	}

	if result.TotalRows > 0 {
		if err = util.JSONBind(comparisonData, result.Rows[0]); err != nil {
			r.logger.Error(ctx, "failed while binding the result value to the struct", logger.Error(err))
			return nil, fmt.Errorf("JSON Binding: %v", err)
		}
	}

	return comparisonData, nil
}
