package paymentRepository

import (
	"context"
	"fmt"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
)

// GetInvestigationSummary retrieves summary statistics for investigation statuses
func (r *PaymentRepository) GetInvestigationSummary(ctx context.Context, opt paymentModel.GetInvestigationSummaryOption) (*paymentModel.InvestigationSummaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetInvestigationSummary")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments")

	query := `
		SELECT
			COALESCE(SUM(IF(reason_type = 'INVESTIGATION_IN_PROCESS', amount, 0)), 0) AS total_in_progress,
			COALESCE(SUM(IF(reason_type = 'INVESTIGATION_SUCCESS', amount, 0)), 0) AS total_success,
			COALESCE(SUM(IF(reason_type = 'INVESTIGATION_FAILED', amount, 0)), 0) AS total_failed,
			COALESCE(MAX(currency), 'IDR') AS currency
		FROM payments
		WHERE merchant_id = ?
			AND reason_type IN ('INVESTIGATION_IN_PROCESS', 'INVESTIGATION_SUCCESS', 'INVESTIGATION_FAILED')
			AND investigation_started_at BETWEEN ? AND ?
			AND created_at >= DATE_SUB(?, INTERVAL 60 DAY)
	`

	args := []any{opt.MerchantID, opt.StartDate, opt.EndDate, opt.StartDate}

	var row paymentModel.SummaryRow
	err := r.db.GetContext(ctx, &row, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when getting investigation summary", logger.Error(err))
		return nil, err
	}

	// Build response
	response := &paymentModel.InvestigationSummaryResponse{
		OnInvestigation: paymentModel.InvestigationSummaryItem{
			TotalAmount: fmt.Sprintf("%.2f", row.TotalInProgress),
			Currency:    row.Currency,
		},
		Success: paymentModel.InvestigationSummaryItem{
			TotalAmount: fmt.Sprintf("%.2f", row.TotalSuccess),
			Currency:    row.Currency,
		},
		Failed: paymentModel.InvestigationSummaryItem{
			TotalAmount: fmt.Sprintf("%.2f", row.TotalFailed),
			Currency:    row.Currency,
		},
	}

	return response, nil
}
