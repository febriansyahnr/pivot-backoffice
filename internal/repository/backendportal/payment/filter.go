package paymentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

// FilterPaymentHistory retrieves paginated payment history records based on the provided filter options.
// It joins payments with payment_methods and account_transactions to provide comprehensive payment information.
//
// The function returns payment records with detailed information including:
// - Payment basic info (UUID, merchant ID, reference ID, amount, status, timestamps)
// - Payment method details (type, channel, processor reference number)
// - Transaction settlement information from account_transactions
// - Payment-specific metadata (virtual account details, QRIS info, credit card data)
//
// Returns:
//   - *commonModel.PaginationResponse: Paginated results with payment history items and metadata
//   - error: Database or processing error, if any
//
// The function uses concurrent queries (errgroup) to fetch total count and payment records simultaneously
// for better performance.
func (r *PaymentRepository) FilterPaymentHistory(ctx context.Context, opt paymentModel.FilterPaymentHistoryOption) (*commonModel.PaginationResponse, error) {
	// Initialize pagination utility (assuming no over-fetch for this repository)
	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig != nil && r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      0,
	}
	if r.appConfig != nil {
		paginationConfig.InitialPageWindow = r.appConfig.InitialPageWindow
	}

	paginationUtil := util.NewPaginationUtility(r.db, r.logger, paginationConfig)

	// Build query components
	queryBuilder := util.QueryBuilder{
		SelectQuery: `SELECT 
		p.uuid, p.merchant_id, p.reference_id, 
		p.recurring_contract_id,
		IFNULL(pm.type, '') AS payment_method,
		CASE 
			WHEN IFNULL(pm.type, '') =  "QRIS" THEN COALESCE(p.metadata->>'$.paymentMethodType', "")
			WHEN IFNULL(pm.type, '') =  "VIRTUAL_ACCOUNT" AND (p.metadata->>'$.snapCore.isClosedAmount' = 'true' AND p.metadata->>'$.snapCore.isSingleUse' = 'true') THEN "CLOSED_DYNAMIC"
			WHEN IFNULL(pm.type, '') =  "VIRTUAL_ACCOUNT" AND (p.metadata->>'$.snapCore.isClosedAmount' = 'true' AND p.metadata->>'$.snapCore.isSingleUse' = 'false') THEN "CLOSED_STATIC"
			WHEN IFNULL(pm.type, '') =  "VIRTUAL_ACCOUNT" AND (p.metadata->>'$.snapCore.isClosedAmount' = 'false' AND p.metadata->>'$.snapCore.isSingleUse' = 'false') THEN "OPEN_STATIC"
			ELSE ""
		END AS payment_method_type,
		CASE  
			WHEN IFNULL(pm.type, '') =  "CREDIT_CARD" THEN COALESCE(p.metadata->>'$.cardData.cardBrand', "") 
			ELSE IFNULL(pm.bank_name, '')
		END AS channel, 
		CASE  
			WHEN IFNULL(pm.type, '') =  "CREDIT_CARD" THEN IF(p.metadata->>'$.cardData.last4Digit' IS NULL, "", CONCAT("********", p.metadata->>'$.cardData.last4Digit')) 
			ELSE p.processor_reference_number
		END AS processor_reference_number, 
		p.amount, p.currency AS amount_currency, p.status, p.created_at, at2.updated_at, at2.created_at as transaction_created_at,
		CASE
			WHEN p.type <> 'SINGLE' THEN
				COALESCE(metadata->>'$.summaryTransaction.sumPaidAmount', '0')
			WHEN NULLIF(
				metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value',
				''
			) IS NOT NULL THEN
				COALESCE(metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','0')
			WHEN at2.status = 'SUCCESS' THEN
				COALESCE(CAST(at2.credit AS CHAR),'0')
			ELSE
				'0'
		END AS amount_paid,
		IF(at2.status = 'SUCCESS', at2.currency, NULL) AS amount_paid_currency, 
		at2.transaction_timestamp AS paid_at, p.expired_at, IFNULL(p.customer_id, '') AS customer_id,
		IF(IFNULL(pm.type, '') =  'VIRTUAL_ACCOUNT', IFNULL(metadata->>'$.snapCore.number', ''), '  -') AS virtual_account_no,
		IF(IFNULL(pm.type, '') =  'VIRTUAL_ACCOUNT', IFNULL(metadata->>'$.snapCore.accountName', ''), '  -') AS virtual_account_name,
		IF(IFNULL(pm.type, '') =  'QRIS', IFNULL(metadata->>'$.snapCore.merchantName', ''), '  -') AS qris_merchant_name,
		IF(IFNULL(pm.type, '') =  'QRIS', IFNULL(metadata->>'$.snapCore.qrUrl', ''), '  -') AS qris_url,
		IFNULL(LEFT(p.reason_type, 13) = 'INVESTIGATION', false) AS has_investigation,
		IF(IFNULL(pm.type, '') =  'CREDIT_CARD', IFNULL(at2.additional_info->>'$.methodDetail.card.binInformations.cardType', ''), '') AS credit_card_type,
		IF(IFNULL(pm.type, '') =  'CREDIT_CARD', IFNULL(at2.additional_info->>'$.methodDetail.card.binInformations.issuingBank', ''), '  -') AS credit_card_issuer_bank,
		IF(IFNULL(pm.type, '') = 'CREDIT_CARD', CONCAT( IFNULL(at2.additional_info->>"$.methodDetail.card.first6", ''), '******', IFNULL(at2.additional_info->>"$.methodDetail.card.last4", '')),'  -') AS credit_card_number,
		IF(IFNULL(pm.type, '') = 'CREDIT_CARD', CONCAT( IFNULL(at2.additional_info->>"$.methodDetail.card.expYear", ''), '/', IFNULL(at2.additional_info->>"$.methodDetail.card.expMonth", '')),'  -') AS credit_card_expiry,
		COALESCE(IF(pm.type='CREDIT_CARD', at2.additional_info->>"$.methodDetail.card.midInfo.mid", ''), '') AS mid,
		COALESCE(IF(pm.type='CREDIT_CARD', at2.additional_info->>"$.methodDetail.card.midInfo.type", ''), '') AS mid_type,
		COALESCE(IF(pm.type='CREDIT_CARD', at2.additional_info->>"$.methodDetail.card.midInfo.acquirer", ''), '') AS mid_acquirer,
		COALESCE(at2.settlement_model, 'AGGREGATOR') as settlement_model,
		IFNULL(JSON_LENGTH(metadata->>'$.splitRoutingConfigurations') > 0 AND JSON_CONTAINS_PATH(metadata->>'$.splitRoutingConfigurations[0]', 'one', '$.transferId'), false) AS has_split_routing,
		r.created_at AS refund_date,
		r.amount AS refund_amount,
		r.status AS refund_status,
		refund_agg.total_refunded_amount,
		refund_agg.refund_count,
		IF (p.created_from = 'MERCHANT_PORTAL', IFNULL(metadata->>'$.shortPaymentUrl',p.payment_url), '') as payment_url
		FROM payments p 
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
		LEFT JOIN account_transactions at2 ON at2.reference_id = p.uuid AND at2.type = '` + constant.TypePayment + `' 
		AND (p.type = 'SINGLE' OR at2.created_at = (SELECT MAX(created_at) FROM account_transactions WHERE reference_id = p.uuid AND type = '` + constant.TypePayment + `'))
		LEFT JOIN (
			SELECT payment_id, created_at, amount, status, 
			       ROW_NUMBER() OVER (PARTITION BY payment_id ORDER BY created_at DESC) as rn
			FROM refunds
		) r ON r.payment_id = p.uuid AND r.rn = 1
		LEFT JOIN (
			SELECT payment_id,
			       SUM(amount) AS total_refunded_amount,
			       COUNT(*) AS refund_count
			FROM refunds
			WHERE status = 'SUCCESS'
			GROUP BY payment_id
		) refund_agg ON refund_agg.payment_id = p.uuid`,
		CountQuery: `SELECT COUNT(p.uuid) 
		FROM payments p 
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
		LEFT JOIN account_transactions at2 ON at2.reference_id = p.uuid AND at2.type = '` + constant.TypePayment + `' 
		AND (p.type = 'SINGLE' OR at2.created_at = (SELECT MAX(created_at) FROM account_transactions WHERE reference_id = p.uuid AND type = '` + constant.TypePayment + `'))`,
	}

	// Build filter conditions
	conditions, args := r.buildPaymentHistoryConditions(opt)
	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	// Handle naming convention for sorting
	sortBy := opt.SortBy
	switch opt.SortBy {
	case "createdAt":
		sortBy = "created_at"
	case "amountPaid":
		sortBy = "amount_paid"
	case "paymentDate":
		sortBy = "at2.created_at"
	}

	// Build sort configuration
	sortConfig := util.SortConfig{
		DefaultSort: "p.created_at DESC",
		SortBy:      sortBy,
		Sort:        opt.Sort,
	}

	// Validate pagination parameters
	page := int64(opt.Page)
	perPage := int64(opt.PerPage)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	// Data destination
	data := make([]paymentModel.PaymentHistoryItem, 0)

	// Data transformer
	dataTransformer := func(dest interface{}) interface{} {
		typedData := dest.(*[]paymentModel.PaymentHistoryItem)
		result := *typedData

		for i := range result {
			item := &result[i]
			// Build RefundInfo if there are refunds
			if item.RefundCount != nil && *item.RefundCount > 0 && item.TotalRefundedAmount != nil {
				refundType := constant.RefundTypeNone
				// Compare total refunded amount with payment amount to determine refund type
				if item.Amount != "" && *item.TotalRefundedAmount != "" {
					paymentAmount, errPay := decimal.NewFromString(item.Amount)
					refundedAmount, errRef := decimal.NewFromString(*item.TotalRefundedAmount)
					if errPay == nil && errRef == nil {
						// If refunded amount >= payment amount, consider it FULL
						if refundedAmount.GreaterThanOrEqual(paymentAmount) {
							refundType = constant.RefundTypeFull
						} else if refundedAmount.GreaterThan(decimal.Zero) {
							refundType = constant.RefundTypePartial
						}
					}
				}
				item.RefundInfo = &paymentModel.RefundInfo{
					Type:                refundType,
					TotalRefundedAmount: *item.TotalRefundedAmount,
					RefundCount:         *item.RefundCount,
				}
			}
		}

		return result
	}

	return paginationUtil.GetPaginatedList(
		ctx,
		queryBuilder,
		filterResult,
		sortConfig,
		page,
		perPage,
		&data,
		dataTransformer,
	)
}

// buildPaymentHistoryConditions builds WHERE conditions for payment history filtering
func (r *PaymentRepository) buildPaymentHistoryConditions(opt paymentModel.FilterPaymentHistoryOption) (conditions []string, args []interface{}) {
	conditions = append(conditions, "p.merchant_id = ?")
	args = append(args, opt.MerchantID)

	conditions = append(conditions, "p.type IN ('', 'MULTIPLE', 'SINGLE')")

	if opt.ReferenceID != "" {
		conditions = append(conditions, "(p.reference_id LIKE ? OR p.customer_id = ? OR p.uuid = ?)")
		args = append(args, "%"+opt.ReferenceID+"%", opt.ReferenceID, opt.ReferenceID)
	}

	if opt.PaymentMethod != "" {
		conditions = append(conditions, "pm.type = ?")
		args = append(args, opt.PaymentMethod)
	}

	if opt.Status != "" {
		statuses := strings.Split(opt.Status, ",")

		hasRefundedFilter := util.Contains(statuses, constant.PaymentStatusRefunded)
		if hasRefundedFilter {
			filteredStatuses := make([]string, 0)
			for _, s := range statuses {
				if s != constant.PaymentStatusRefunded {
					filteredStatuses = append(filteredStatuses, s)
				}
			}
			statuses = filteredStatuses
		}

		if util.Contains(statuses, constant.UnifiedPaymentSessionStatusPaid) {
			statuses = append(statuses, constant.StatusSuccess)
		}

		if util.Contains(statuses, constant.UnifiedPaymentSessionStatusProcessing) {
			statuses = append(statuses, constant.StatusPending)
		}

		var statusConditions []string

		if len(statuses) > 0 {
			statusQuery, statusArgs, _ := sqlx.In("p.status IN (?)", statuses)
			statusConditions = append(statusConditions, statusQuery)
			args = append(args, statusArgs...)
		}

		if hasRefundedFilter {
			statusConditions = append(statusConditions, "refund_agg.refund_count > 0")
		}

		if len(statusConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(statusConditions, " OR ")+")")
		}
	}

	if opt.SettlementModel != "" {
		if opt.SettlementModel == constant.PaymentMethodChannelTypeAggregator {
			conditions = append(conditions, "(at2.settlement_model = ? OR at2.settlement_model IS NULL)")
		} else {
			conditions = append(conditions, "at2.settlement_model = ?")
		}
		args = append(args, opt.SettlementModel)
	}

	if !opt.StartDate.IsZero() || !opt.EndDate.IsZero() {
		dateRange := []string{}
		if !opt.StartDate.IsZero() {
			dateRange = append(dateRange, "p.created_at >= ?")
			args = append(args, opt.StartDate)
		}
		if !opt.EndDate.IsZero() {
			dateRange = append(dateRange, "p.created_at <= ?")
			args = append(args, opt.EndDate)
		}

		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(dateRange, " AND ")))
	}

	if !opt.PaymentStartDate.IsZero() || !opt.PaymentEndDate.IsZero() {
		dateRange := []string{}
		if !opt.PaymentStartDate.IsZero() {
			dateRange = append(dateRange, "at2.created_at >= ?")
			args = append(args, opt.PaymentStartDate)
		}
		if !opt.PaymentEndDate.IsZero() {
			dateRange = append(dateRange, "at2.created_at <= ?")
			args = append(args, opt.PaymentEndDate)
		}

		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(dateRange, " AND ")))
	}

	return conditions, args
}

// GetPaymentIDsExpiringToday returns payment IDs that will expire from start to end
// logic that used on date rang is >= start and < end
func (r *PaymentRepository) GetExpiringPayments(ctx context.Context, start time.Time, end time.Time) ([]*paymentModel.ExpiringPayment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentsExpiringToday")
	defer segment.End()

	var (
		payments []*paymentModel.ExpiringPayment
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	// use the '<' to provide time precision
	// gather data from 01:00 until 00:59:59
	query := `
			SELECT 
				p.uuid,
				p.merchant_id,
				p.expired_at
			FROM payments p
			WHERE p.expired_at >= ? AND p.expired_at < ? AND 
			p.status NOT IN ('EXPIRED', 'SUCCESS', 'CANCELLED', 'FAILED', 'VOID')
			ORDER BY p.created_at ASC;
	`

	err := r.db.SelectContext(ctx, &payments, query, start, end)
	if err != nil && err == sql.ErrNoRows {
		r.logger.Info(ctx, "no payments expiring today")
		return nil, nil
	}

	if err != nil {
		r.logger.Error(ctx, "error when getting payments expiring today", logger.Error(err))
		return nil, err
	}

	return payments, nil
}
