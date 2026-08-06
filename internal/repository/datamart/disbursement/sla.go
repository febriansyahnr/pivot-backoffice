package disbursementDatamart

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const MetricNameSuccessSLA = "Success SLA"

func (r *repository) GetSLAMetrics(
	ctx context.Context,
	filter disbursementModel.GetDisbursementInsightFilter,
) (*disbursementModel.SLAMetrics, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/datamart/disbursement/GetSLAMetrics")
	defer segment.End()

	if r.db == nil {
		r.logger.Info(ctx, "BigQuery service not configured, returning default SLA metrics")
		return r.getDefaultSLAMetrics(), nil
	}

	result, err := r.getMetricsFromBigQuery(ctx, filter, MetricNameSuccessSLA)
	if err != nil {
		r.logger.Error(ctx, "failed to execute BigQuery SLA metrics query", logger.Error(err))
		return r.getDefaultSLAMetrics(), nil
	}

	metrics := r.getDefaultSLAMetrics()

	if len(result.Rows) == 0 {
		return metrics, nil
	}

	var dailyBreakdown []disbursementModel.DailySLAMetric
	var processingTimeSum float64

	for _, row := range result.Rows {
		date := r.getStringFromInterface(row["transaction_date"])
		processingTime := r.getFloat64FromInterface(row["metric_numerator"])

		dailyMetric := disbursementModel.DailySLAMetric{
			Date:                         date,
			AverageProcessingTimeMinutes: processingTime,
		}

		dailyBreakdown = append(dailyBreakdown, dailyMetric)
		processingTimeSum += processingTime
	}

	if len(dailyBreakdown) > 0 {
		metrics.AverageProcessingTimeMinutes = processingTimeSum / float64(len(dailyBreakdown))
	}

	sort.Slice(dailyBreakdown, func(i, j int) bool {
		return dailyBreakdown[i].Date > dailyBreakdown[j].Date
	})

	metrics.DailyBreakdown = dailyBreakdown
	return metrics, nil
}

func (r *repository) getDefaultSLAMetrics() *disbursementModel.SLAMetrics {
	return &disbursementModel.SLAMetrics{
		AverageProcessingTimeMinutes: 0.0,
		DailyBreakdown:               []disbursementModel.DailySLAMetric{},
	}
}

func (r *repository) getFloat64FromInterface(val any) float64 {
	if val == nil {
		return 0.0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return 0.0
}

func (r *repository) QueryDisbursementSLAComparison(ctx context.Context, request disbursementModel.QueryDisbursementSLAComparisonRequest) (*disbursementModel.ComparisonData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/datamart/disbursement/QueryDisbursementSLAComparison")
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

	rawQuery := fmt.Sprintf(`WITH disbursement_sla_by_range AS (
		SELECT
			AVG(IF(transaction_date BETWEEN @prev_start_date AND @prev_end_date, metric_numerator, NULL)) AS prev_avg_processing_time,
			AVG(IF(transaction_date BETWEEN @current_start_date AND @current_end_date, metric_numerator, NULL)) AS curr_avg_processing_time
		FROM 
			%s
		WHERE
			merchant_id = @merchant_id
			AND metric_name = 'Success SLA'
			AND transaction_date BETWEEN @prev_start_date AND @current_end_date
	) SELECT 
		COALESCE(prev_avg_processing_time, 0) AS `+"`previous`"+`,
		COALESCE(curr_avg_processing_time, 0) AS `+"`current`"+`,
		ROUND(COALESCE(curr_avg_processing_time, 0) - COALESCE(prev_avg_processing_time, 0), 2) AS `+"`difference`"+`
	FROM disbursement_sla_by_range;`, r.getTableName())

	result, err := r.db.ExecuteQueryWithParams(ctx, rawQuery, params)
	if err != nil {
		r.logger.Error(ctx, "failed to execute disbursement SLA comparison query", logger.Error(err))
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
