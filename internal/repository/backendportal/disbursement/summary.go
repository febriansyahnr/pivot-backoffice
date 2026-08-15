package disbursementRepository

import (
	"context"
	"time"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursementDashboard"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

var (
	timeLoc, _     = time.LoadLocation(constant.TimeLoc)
	currentTime    = time.Now().In(timeLoc)
	beginningOfDay = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, timeLoc)
	defaultSummary = disbursementDashboardModel.SummaryTransactionDTO{
		Count: 0,
		Sum:   0,
	}
)

// GetSummaryAll will calculate all transaction based on the filter param
func (r *DisbursementRepository) GetSummaryAll(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryAll")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND (created_at > ? AND created_at <= ?)"
	args := []interface{}{filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count all today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

// GetSummaryAll will calculate all Success transaction based on the filter param where the disbursement is approved and the transaction was success
func (r *DisbursementRepository) GetSummarySuccess(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummarySuccess")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.amount),0) as 'sum'
		FROM disbursements d
		INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.merchant_id = ? AND (d.created_at > ? AND d.created_at <= ?) AND d.status = ? AND t.status = ?`

	args := []interface{}{constant.TypeDisbursement, filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusApproved, constant.StatusSuccess}
	if filter.IsXbPayout {
		queryCount += " AND d.currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND d.currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count success today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

// GetSummaryFailed will calculate all Failed transaction based on the filter param where the disbursement was approved and the transaction was failed
func (r *DisbursementRepository) GetSummaryFailed(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryFailed")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.amount),0) as 'sum'
		FROM disbursements d
		INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.merchant_id = ? AND (d.created_at > ? AND d.created_at <= ?) AND d.status = ? AND t.status = ?`

	args := []interface{}{constant.TypeDisbursement, filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusApproved, constant.StatusFailed}
	if filter.IsXbPayout {
		queryCount += " AND d.currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND d.currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count failed today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

// GetSummaryFailed will calculate all Failed transaction based on the filter param where the disbursement was approved and the transaction was pending
func (r *DisbursementRepository) GetSummaryInProgress(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryInProgress")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.amount),0) as 'sum'
		FROM disbursements d
		LEFT JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.merchant_id = ? AND (d.created_at > ? AND d.created_at <= ?) AND d.status = ? AND t.status = ?`

	args := []interface{}{constant.TypeDisbursement, filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusApproved, constant.StatusPending}
	if filter.IsXbPayout {
		queryCount += " AND d.currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND d.currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(
		ctx,
		&summary,
		queryCount,
		args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count in progress today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryWaitingToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ?"
	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusWaiting}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummarySingleWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummarySingleWaitingToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ? AND bulk_id IS NULL"

	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusWaiting}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting today single disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryBulkWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryBulkWaitingToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ? AND bulk_id IS NOT NULL"
	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusWaiting}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting today bulk disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryWaitingForTopUpToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ? AND reason_type = ?"
	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusApproved, constant.DisbursementReasonTypeInsufficientBalance}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting for top up today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummarySingleWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummarySingleWaitingForTopUpToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ? AND reason_type = ? AND bulk_id IS NULL"
	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusApproved, constant.DisbursementReasonTypeInsufficientBalance}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting for top up today single disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryBulkWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryBulkWaitingForTopUpToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND created_at > ? AND status = ? AND reason_type = ? AND bulk_id IS NOT NULL"
	args := []interface{}{filter.MerchantID, beginningOfDay, constant.DisbursementStatusApproved, constant.DisbursementReasonTypeInsufficientBalance}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count waiting for top up today bulk disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

// GetSummaryRejected will return summary of reject disbursement based on filter date
func (r *DisbursementRepository) GetSummaryRejected(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryRejected")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND (created_at > ? AND created_at <= ?) AND status = ?"
	args := []interface{}{filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusRejected}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count rejected today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

// GetSummaryApproved will return summary of approved disbursement based on filter date
func (r *DisbursementRepository) GetSummaryApproved(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryApproved")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := "SELECT COUNT(uuid) as 'count', COALESCE(SUM(amount),0) as 'sum' FROM disbursements WHERE merchant_id = ? AND (created_at > ? AND created_at <= ?) AND status = ?"
	args := []interface{}{filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusApproved}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count rejected today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummarySuccessByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummarySuccessByBulkID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.total_amount),0) as 'sum'
		FROM disbursements d
		INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.bulk_id = ? AND d.status = ? AND t.status = ?`

	err := r.db.GetContext(ctx, &summary, queryCount, constant.TypeDisbursement, bulkID, constant.DisbursementStatusApproved, constant.StatusSuccess)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count success disbursements by bulkID", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryFailedByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryFailedToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.total_amount),0) as 'sum'
		FROM disbursements d
		INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.bulk_id = ? AND d.status = ? AND t.status = ?`

	err := r.db.GetContext(ctx, &summary, queryCount, constant.TypeDisbursement, bulkID, constant.DisbursementStatusApproved, constant.StatusFailed)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count failed disbursements by bulkID", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryCancelledByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryCancelledByBulkID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.total_amount),0) as 'sum'
		FROM disbursements d
		WHERE d.bulk_id = ? AND d.status = ? AND d.reason_type = ?`

	err := r.db.GetContext(ctx, &summary, queryCount, bulkID, constant.DisbursementStatusApproved, constant.DisbursementReasonTypeCancelled)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count cancelled disbursements by bulkID", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) SummaryPendingByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SummaryInProgressToday")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
			COUNT(d.uuid) as 'count',
			COALESCE(SUM(d.total_amount),0) as 'sum'
		FROM disbursements d
		LEFT JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.bulk_id = ?
		AND 
		(
			d.status NOT IN ('` + constant.DisbursementStatusApproved + `', '` + constant.DisbursementStatusRejected + `')
			OR (
				d.status = '` + constant.DisbursementStatusApproved + `' 
				AND (t.status NOT IN ('` + constant.StatusSuccess + `', '` + constant.StatusFailed + `') OR t.uuid IS NULL)
				AND COALESCE(d.reason_type, '') NOT IN ('` + constant.DisbursementReasonTypeCancelled + `')
			)
		)
		`

	err := r.db.GetContext(
		ctx,
		&summary,
		queryCount,
		constant.TypeDisbursement,
		bulkID)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count pending today disbursements", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) GetSummaryByReasonType(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter, transactionStatus string) ([]disbursementDashboardModel.SummaryTransactionByReasonType, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryByReasonType")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summaries []disbursementDashboardModel.SummaryTransactionByReasonType

	queryCount := `SELECT 
			COALESCE(NULLIF(t.reason_type, ''), 'OTHER') as reason_type,
			COUNT(t.uuid) as 'count',
			COALESCE(SUM(d.amount),0) as 'sum'
		FROM disbursements d
		INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = ?
		WHERE d.merchant_id = ? AND (d.created_at > ? AND d.created_at <= ?) AND d.status = ? AND t.status = ?`

	args := []any{constant.TypeDisbursement, filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, constant.DisbursementStatusApproved, transactionStatus}
	if filter.IsXbPayout {
		queryCount += " AND d.currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND d.currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	queryCount += " GROUP BY reason_type ORDER BY sum DESC"

	err := r.db.SelectContext(ctx, &summaries, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count disbursements by reason type", logger.Error(err))
		return []disbursementDashboardModel.SummaryTransactionByReasonType{}, nil
	}

	return summaries, nil
}

func (r *DisbursementRepository) GetSummaryByDisbursementStatus(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter, disbursementStatus string) disbursementDashboardModel.SummaryTransactionDTO {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryByDisbursementStatus")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var summary disbursementDashboardModel.SummaryTransactionDTO

	queryCount := `SELECT 
		COUNT(uuid) as 'count', 
		COALESCE(SUM(amount),0) as 'sum' 
	FROM disbursements 
	WHERE merchant_id = ? 
		AND (created_at > ? AND created_at <= ?) 
		AND status = ?`

	args := []any{filter.MerchantID, filter.InsightStartDate, filter.InsightEndDate, disbursementStatus}
	if filter.IsXbPayout {
		queryCount += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		queryCount += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &summary, queryCount, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count disbursements by status", logger.Error(err))
		return defaultSummary
	}

	return summary
}

func (r *DisbursementRepository) GetActionTransactionSummary(ctx context.Context, merchantId string, disbursementIds []string) (*disbursementModel.ActionTransactionSummary, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetActionTransactionSummary")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements, merchants")

	rawQuery := `SELECT COUNT(uuid) AS total, IFNULL(SUM(amount), 0) AS total_amount FROM disbursements d WHERE merchant_id = ? AND uuid IN (?);`

	query, args, err := r.db.In(rawQuery, merchantId, disbursementIds)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query) // Formatting appropriate to the driver used

	result := &disbursementModel.ActionTransactionSummary{}
	return result, r.db.GetContext(ctx, result, query, args...)
}
