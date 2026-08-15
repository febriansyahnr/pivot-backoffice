package disbursementRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursementDashboard"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *DisbursementRepository) CountByIDsAndMerchantID(ctx context.Context, ids []string, merchantID string) (count int) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountByIDsAndMerchantID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	rawQuery := "SELECT COUNT(uuid) FROM disbursements WHERE uuid IN (?) AND merchant_id = ?"

	query, args, err := r.db.In(rawQuery, ids, merchantID)
	if err != nil {
		r.pdkLogger.Error(ctx, "failed parsing query (r.db.In)", logger.Error(err))
		return
	}
	query = r.db.Rebind(query) // Formatting appropriate to the driver used

	if err = r.db.GetContext(ctx, &count, query, args...); err != nil {
		r.pdkLogger.Error(ctx, "error when count disbursements", logger.Error(err))
	}
	return
}

func (r *DisbursementRepository) CountByBulkID(ctx context.Context, bulkID string) int {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountByBulkID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var totalItems = 0

	queryCount := "SELECT COUNT(uuid) as totalItems FROM disbursements WHERE bulk_id = ?"
	err := r.db.GetContext(ctx, &totalItems, queryCount, bulkID)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count disbursements", logger.Error(err))
		totalItems = 0
	}

	return totalItems
}

func (r *DisbursementRepository) CountWaitingSingleDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountWaitingSingleDisbursement")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var totals disbursementDashboardModel.SummaryTransactionDTO

	query := `
		SELECT 
			COUNT(uuid) AS count,
			COALESCE(SUM(amount), 0) AS sum
		FROM 
			disbursements 
		WHERE 
			merchant_id = ? 
			AND status = ? 
			AND bulk_id IS NULL
	`
	args := []interface{}{filter.MerchantID, constant.DisbursementStatusWaiting}
	if filter.IsXbPayout {
		query += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		query += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}
	err := r.db.GetContext(ctx, &totals, query, args...)

	if err != nil {
		r.pdkLogger.Error(ctx, "error when counting waiting single disbursements", logger.Error(err))
		return disbursementDashboardModel.SummaryTransactionDTO{}, err
	}

	return totals, nil
}

// CountWaitingBulkDisbursement retrieves the count and total amount of waiting bulk disbursements.
func (r *DisbursementRepository) CountWaitingBulkDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountWaitingBulkDisbursement")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	var totals disbursementDashboardModel.SummaryTransactionDTO

	query := `
		SELECT 
			COUNT(DISTINCT bd.uuid) AS count,
			COALESCE(SUM(d.amount), 0) AS sum
		FROM 
			bulk_disbursements bd
		LEFT JOIN 
			disbursements d ON d.bulk_id = bd.uuid
		WHERE 
			bd.merchant_id = ? 
			AND bd.status = ?
	`
	args := []interface{}{filter.MerchantID, constant.DisbursementStatusWaiting}
	err := r.db.GetContext(ctx, &totals, query, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when counting waiting bulk disbursements", logger.Error(err))
		return disbursementDashboardModel.SummaryTransactionDTO{}, err
	}

	return totals, nil
}

// CountPendingSingleDisbursement retrieves the count and total amount of pending single disbursements.
func (r *DisbursementRepository) CountPendingSingleDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountPendingSingleDisbursement")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var totals disbursementDashboardModel.SummaryTransactionDTO

	query := `
		SELECT 
			COUNT(uuid) AS count,
			COALESCE(SUM(amount), 0) AS sum
		FROM 
			disbursements 
		WHERE 
			merchant_id = ? 
			AND status = ? 
			AND reason_type = ? 
			AND bulk_id IS NULL
	`
	args := []interface{}{filter.MerchantID, constant.DisbursementStatusApproved, constant.DisbursementReasonTypeInsufficientBalance}
	if filter.IsXbPayout {
		query += " AND currency != ?"
		args = append(args, constant.CurrencyIDR)
	} else {
		query += " AND currency = ?"
		args = append(args, constant.CurrencyIDR)
	}

	err := r.db.GetContext(ctx, &totals, query, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when count pending disbursements", logger.Error(err))
		return disbursementDashboardModel.SummaryTransactionDTO{}, err
	}

	return totals, nil
}

// CountPendingBulkDisbursement retrieves the count and total amount of pending bulk disbursements.
func (r *DisbursementRepository) CountPendingBulkDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountPendingBulkDisbursement")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	var totals disbursementDashboardModel.SummaryTransactionDTO

	query := `
		SELECT 
			COUNT(DISTINCT bd.uuid) AS count,
			COALESCE(SUM(d.amount), 0) AS sum
		FROM 
			bulk_disbursements bd
		LEFT JOIN 
			disbursements d ON d.bulk_id = bd.uuid
		WHERE 
			bd.merchant_id = ? 
			AND bd.status = ?
	`
	args := []interface{}{filter.MerchantID, constant.BulkDisbursementStatusPending}
	err := r.db.GetContext(ctx, &totals, query, args...)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when counting pending bulk disbursements", logger.Error(err))
		return disbursementDashboardModel.SummaryTransactionDTO{}, err
	}

	return totals, nil
}

func (r *DisbursementRepository) CountByMerchantAndReference(ctx context.Context, merchantID, referenceID string) int {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountByMerchantAndReference")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var totalItems = 0

	// note: reference_id lookup is case insensitive
	queryCount := "SELECT COUNT(uuid) as totalItems FROM disbursements WHERE merchant_id = ? AND reference_id = ?"
	err := r.db.GetContext(ctx, &totalItems, queryCount, merchantID, referenceID)
	if err != nil {
		r.pdkLogger.Error(
			ctx,
			"error when count disbursements by referenceID",
			logger.Error(err),
			logger.String("merchantId", merchantID),
			logger.String("referenceId", referenceID),
		)
		totalItems = 0
	}

	return totalItems
}

func (r *DisbursementRepository) CountStatusInProgressByBulkID(ctx context.Context, bulkID string) int {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/CountStatusInProgressByBulkID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var totalItems = 0

	queryCount := `
		SELECT COUNT(d.uuid) from disbursements d
		LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id
		WHERE 
			d.bulk_id = ?
			AND 
			(
				d.status NOT IN ('` + constant.DisbursementStatusApproved + `', '` + constant.DisbursementStatusRejected + `')
				OR (
					d.status = '` + constant.DisbursementStatusApproved + `' 
					AND (t.status NOT IN ('` + constant.StatusSuccess + `', '` + constant.StatusFailed + `') OR t.uuid IS NULL)
				)
			)
	`
	err := r.db.GetContext(ctx, &totalItems, queryCount, bulkID)
	if err != nil {
		r.pdkLogger.Error(
			ctx,
			"error when count status in progress disbursements by bulkID",
			logger.Error(err),
			logger.String("bulkID", bulkID),
		)
		totalItems = 0
	}

	return totalItems
}
