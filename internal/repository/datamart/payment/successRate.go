package paymentDatamart

import (
	"context"
	"fmt"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
)

// QueryPaymentSuccessRateComparison is a function that calculates the payment transaction success rate for the current period and the previous period,
// and the difference between the two. Date inputs use a date-only format (YYYY-MM-DD) and are interpreted in the Asia/Jakarta time zone.
func (r *repository) QueryPaymentSuccessRateComparison(ctx context.Context, request model.QueryPaymentSuccessRateComparisonRequest) (*model.PaymentSuccessRateComparison, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/datamart/payment/QueryPaymentSuccessRateComparison")
	defer segment.End()

	params := map[string]any{
		"merchant_id":        request.MerchantId,
		"prev_start_date":    request.PrevRange.StartDate,
		"prev_end_date":      request.PrevRange.EndDate,
		"current_start_date": request.CurrentRange.StartDate,
		"current_end_date":   request.CurrentRange.EndDate,
	}

	rawQuery := fmt.Sprintf(`WITH payment_grouping_by_range AS (
		SELECT
			COALESCE(SUM(IF(transaction_date BETWEEN @prev_start_date AND @prev_end_date, metric_numerator, 0)), 0) AS prev_payment_success_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @prev_start_date AND @prev_end_date, metric_denominator, 0)), 0) AS prev_payment_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @current_start_date AND @current_end_date, metric_numerator, 0)), 0) AS curr_payment_success_count,
			COALESCE(SUM(IF(transaction_date BETWEEN @current_start_date AND @current_end_date, metric_denominator, 0)), 0) AS curr_payment_count
		FROM 
			%s
		WHERE
			merchant_id = @merchant_id
			AND metric_name = 'Payment Success Rate'
			AND transaction_date BETWEEN @prev_start_date AND @current_end_date
	), payment_success_rate AS (
		SELECT 
			ROUND(SAFE_DIVIDE(prev_payment_success_count, prev_payment_count) * 100, 2) AS previous_success_rate,
			ROUND(SAFE_DIVIDE(curr_payment_success_count, curr_payment_count) * 100, 2) AS current_success_rate
		FROM payment_grouping_by_range
	) SELECT 
		@merchant_id AS merchantId,
		previous_success_rate AS previousSuccessRate, 
		current_success_rate AS currentSuccessRate, 
		ROUND((current_success_rate - previous_success_rate), 2) AS differenceRate
	FROM payment_success_rate;`, r.config.PaymentSuccessMetricsTable)

	result, err := r.db.ExecuteQueryWithParams(ctx, rawQuery, params)
	if err != nil {
		r.logger.Error(ctx, "failed to execute payment success rate query", logger.Error(err))
		return nil, err
	}

	successRate := &model.PaymentSuccessRateComparison{}
	if result.TotalRows > 0 {
		if err = util.JSONBind(successRate, result.Rows[0]); err != nil {
			r.logger.Error(ctx, "failed while binding the result value to the struct", logger.Error(err))
			return nil, fmt.Errorf("JSON Binding: %v", err)
		}
	}
	return successRate, nil
}
