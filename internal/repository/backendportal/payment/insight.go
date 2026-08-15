package paymentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
)

// GetTodayPaymentStatusInsight return pointer of paymentModel.PaymentInsightItem (total payment and total amount)
// but it only contain the total and the amount value, the currency will be assigned using account currency
// the insight will be gathered in UTC timezone
func (r *PaymentRepository) GetTodayPaymentStatusInsight(ctx context.Context, opt paymentModel.PaymentInsightOption) (*paymentModel.PaymentInsightItem, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetTodayPaymentStatusInsight")
	defer segment.End()

	var (
		loc, _      = time.LoadLocation(constant.TimeLoc)
		now         = time.Now().In(loc)
		startOfDay  = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC() // need to convert the timezone because the payment creation using UTC
		queryResult paymentModel.PaymentInsightQueryResult
		insight     *paymentModel.PaymentInsightItem
	)

	query := `
		SELECT COUNT(*) AS total_payment, SUM(p.amount) AS total_amount
		FROM payments p
		WHERE p.merchant_id = ? AND p.created_at >= ? AND status = ?
		GROUP BY p.merchant_id, p.status
	`

	err := r.db.GetContext(ctx, &queryResult, query, opt.MerchantID, startOfDay, opt.Status)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Error(ctx, "error when count failed today disbursements", logger.Error(err))
		return insight, err
	}

	insight = &paymentModel.PaymentInsightItem{
		Total: queryResult.Total,
		TotalAmount: commonModel.Amount{
			Value: strconv.FormatFloat(queryResult.TotalAmount, 'f', 2, 64),
		},
	}

	return insight, nil
}

func (r *PaymentRepository) GetPaymentDashboardInsights(ctx context.Context, request paymentModel.GetPaymentDashboardInsightRequest) (*paymentModel.PaymentDashboardInsights, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetPaymentDashboardInsights")
	defer segment.End()

	args := []any{
		request.StartDate, request.MerchantId, request.StartDate, request.EndDate, request.StartDate,
		request.StartDate, request.MerchantId, request.StartDate, request.EndDate, request.StartDate,
	}
	rawQuery := `WITH aggregate AS (
			SELECT
				SUM(IF(at.type = 'PAYMENT' AND at.status = 'PENDING', 1, 0)) AS waiting_for_capture_count,
				SUM(IF(at.type = 'PAYMENT' AND at.status = 'SUCCESS', 1, 0)) AS paid_count,
				SUM(IF(at.type = 'PAYMENT' AND at.status = 'SUCCESS', at.credit - IF(rf.status = 'SUCCESS', rf.amount, 0), 0)) AS paid_total,
				SUM(IF(at.type = 'REFUND' AND at.status = 'SUCCESS', 1, 0)) AS refunded_count,
				SUM(IF(at.type = 'REFUND' AND at.status = 'SUCCESS', at.debit, 0)) AS refunded_total,
				SUM(IF(at.type = 'PAYMENT' AND at.status = 'FAILED', 1, 0)) AS failed_count,
				SUM(IF(at.type = 'PAYMENT' AND at.status = 'FAILED', at.credit, 0)) AS failed_total,
				SUM(IF(at.type = 'REFUND' AND at.status = 'FAILED', 1, 0)) AS failed_refund_count
			FROM account_transactions at
			LEFT JOIN payments p ON p.uuid = at.reference_id AND p.created_at >= ?
			LEFT JOIN refunds rf ON rf.payment_id = at.reference_id
			WHERE 
				at.merchant_id = ?
				AND (at.created_at >= ? AND at.created_at <= ?) AND at.updated_at >= ?
				AND at.type IN ('PAYMENT', 'REFUND') AND (p.status IS NULL OR p.status != 'EXPIRED')
		),
		failure_breakdown AS (
			SELECT
				CASE 
					WHEN NULLIF(at.additional_info->>'$.failureCode', '') IS NOT NULL THEN at.additional_info->>'$.failureCode'
					WHEN at.reason_type IS NOT NULL THEN at.reason_type 
					ELSE 'OTHERS'
				END AS failure_code, COUNT(at.uuid) AS total
			FROM account_transactions at
			JOIN payments p ON p.uuid = at.reference_id AND p.created_at >= ?
			WHERE 
				at.merchant_id = ?
				AND (at.created_at >= ? AND at.created_at <= ?) AND at.updated_at >= ?
				AND p.status != 'EXPIRED' AND at.status = 'FAILED'
			GROUP BY failure_code 
		)
		SELECT 
			IFNULL(agg.waiting_for_capture_count, 0) AS waiting_for_capture_count, 
			IFNULL(agg.paid_count, 0) AS paid_count,
			IFNULL(agg.paid_total, 0) AS paid_total, 
			IFNULL(agg.refunded_count, 0) AS refunded_count,
			IFNULL(agg.refunded_total, 0) AS refunded_total, 
			IFNULL(agg.failed_count, 0) AS failed_count, 
			IFNULL(agg.failed_total, 0) AS failed_total, 
			IFNULL(agg.failed_refund_count, 0) AS failed_refund_count,
			IF(IFNULL(agg.failed_count, 0) > 0, JSON_ARRAYAGG(
				JSON_OBJECT(
					'failureCode', fb.failure_code,
					'count', fb.total,
					'percentage', ROUND((fb.total/agg.failed_count)*100,2)
				)
			), NULL) AS failure_reasons
		FROM aggregate agg
		LEFT JOIN failure_breakdown fb ON 1 = 1
		GROUP BY
			waiting_for_capture_count, paid_count, paid_total, refunded_count, refunded_total, failed_count, failed_total, failed_refund_count;`

	result := &paymentModel.PaymentDashboardInsights{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		r.logger.Error(ctx, "failed while executing aggregate query", logger.Error(err))
		return nil, err
	}
	if result.RawFailureReasons.Valid {
		if err := result.RawFailureReasons.Unmarshal(&result.FailureReasons); err != nil {
			r.logger.Error(ctx, "failed when unmarshal failure reasons", logger.Error(err))
			return nil, fmt.Errorf("json unmarshal: %v", err)
		}
		sort.Slice(result.FailureReasons, func(i, j int) bool {
			return result.FailureReasons[i].Percentage > result.FailureReasons[j].Percentage
		})
	}
	return result, nil
}
