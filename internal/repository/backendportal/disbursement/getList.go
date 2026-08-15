package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *DisbursementRepository) GetList(
	ctx context.Context,
	filter *disbursementModel.GetDisbursementFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/disbursement/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	// Initialize pagination utility
	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig != nil && r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      0,
	}
	if r.appConfig != nil {
		paginationConfig.InitialPageWindow = r.appConfig.InitialPageWindow
	}

	paginationUtil := util.NewPaginationUtility(r.db, r.pdkLogger, paginationConfig)

	// Build query components
	queryBuilder := util.QueryBuilder{
		SelectQuery: `SELECT ` + SelectDisbursementWithTransactionStr + `
	FROM disbursements d
	LEFT JOIN account_transactions t ON d.uuid = t.reference_id AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
	LEFT JOIN users c ON c.uuid = d.created_by
	LEFT JOIN users a ON a.uuid = d.approved_by`,
		CountQuery: `SELECT COUNT(d.uuid) as totalItems FROM disbursements d
		LEFT JOIN account_transactions t ON d.uuid = t.reference_id AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
		LEFT JOIN users c ON c.uuid = d.created_by
		LEFT JOIN users a ON a.uuid = d.approved_by`,
	}

	// Build filter conditions
	conditions, args := buildCondition(filter)
	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	sortBy := filter.SortBy
	switch filter.SortBy {
	case "createdAt":
		sortBy = "d.created_at"
	case "updatedAt":
		sortBy = "d.updated_at"
	}

	// Build sort configuration
	sortConfig := util.SortConfig{
		DefaultSort: "d.created_at DESC",
		SortBy:      sortBy,
		Sort:        filter.Sort,
	}

	// Data destination
	data := make([]*disbursementModel.DisbursementWithTransaction, 0)

	// Data transformer
	dataTransformer := func(dest interface{}) interface{} {
		typedData := dest.(*[]*disbursementModel.DisbursementWithTransaction)
		return BuildRespData(*typedData)
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

func BuildRespData(
	data []*disbursementModel.DisbursementWithTransaction) []*disbursementModel.DisbursementWithTransactionResponse {
	respData := make([]*disbursementModel.DisbursementWithTransactionResponse, len(data))
	for i, disbursementWithTransaction := range data {
		respData[i] = disbursementWithTransaction.DisbursementWithTransactionToResponse()
	}

	return respData
}

func buildCondition(filter *disbursementModel.GetDisbursementFilterRequest) (conditions []string, args []interface{}) {
	if filter.MerchantID != "" {
		conditions = append(conditions, "d.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}

	if filter.UUID != "" {
		conditions = append(conditions, "(d.uuid = ? OR d.reference_id = ?)")
		args = append(args, filter.UUID, filter.UUID)
	}

	if strings.ToUpper(filter.Status) == constant.DisbursementStatusPending {
		conditions = append(conditions, "d.status = ?")
		args = append(args, constant.DisbursementStatusApproved)

		conditions = append(conditions, "d.reason_type = ?")
		args = append(args, constant.DisbursementReasonTypeInsufficientBalance)
	} else if filter.Status != "" {
		conditions = append(conditions, "d.status LIKE ?")
		args = append(args, "%"+filter.Status+"%")
	}

	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "d.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "d.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}

	// Only allowed LOCAL_PAYOUT and INTERNATIONAL_PAYOUT -> Exclude (CARD_FUNDED_PAYOUT)
	conditions = append(conditions, "d.type IN ('LOCAL_PAYOUT', 'INTERNATIONAL_PAYOUT')")

	if filter.Type == constant.DisbursementTypeSingle {
		conditions = append(conditions, "d.bulk_id IS NULL")
	} else if filter.Type == constant.DisbursementTypeBulk {
		conditions = append(conditions, "d.bulk_id IS NOT NULL")
	}

	if filter.BulkID != "" {
		conditions = append(conditions, "d.bulk_id = ?")
		args = append(args, filter.BulkID)
	}

	if filter.Keyword != "" {
		conditions = append(conditions, "(d.beneficiary_account_name LIKE ? OR d.reference_id LIKE ?)")
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	if filter.TransactionStatus != "" {
		switch strings.ToUpper(filter.TransactionStatus) {
		case constant.DisbursementReasonTypeCancelled:
			conditions = append(conditions, "(d.reason_type = 'CANCELLED' OR t.status = ?)")
			args = append(args, filter.TransactionStatus)
		default:
			conditions = append(conditions, "t.status = ?")
			args = append(args, filter.TransactionStatus)
		}
	}

	if filter.IsXbPayout {
		conditions = append(conditions, "d.currency != ?")
		args = append(args, constant.CurrencyIDR)
	} else {
		conditions = append(conditions, "d.currency = ?")
		args = append(args, constant.CurrencyIDR)
	}

	if filter.ReasonType != "" {
		conditions = append(conditions, "d.reason_type = ?")
		args = append(args, filter.ReasonType)
	}

	return conditions, args
}

func (r *DisbursementRepository) GetCardFundedPayoutTransactionList(ctx context.Context, request cardFundedPayoutModel.GetPayoutTransactionListRequest) ([]cardFundedPayoutModel.GetPayoutTransactionListResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetCardFundedPayoutTransactionList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "account_transactions,disbursements")

	rawQuery := `SELECT
		id, IF(first_payment_status != 'PAID', '', trx_id) AS trx_id,
		client_reference_id, vendor_id, vendor_name, bank_code, bank_name, account_number, 
		account_name, remarks, trx_amount, trx_status, trx_reason_type, trx_reason_desc, 
		created_at, approved_at, 
		IF(first_payment_status != 'PAID', NULL, scheduled_at) AS scheduled_at, 
		IF(first_payment_status != 'PAID', NULL, trx_created_at) AS trx_created_at, 
		IF(first_payment_status != 'PAID', NULL, trx_updated_at) AS trx_updated_at,
		merchant_id, init_amount, init_fee, init_total_amount, 
		IF(trx_status IN ('INITIATE', 'FAILED'), 0, trx_amount) AS exec_amount,
		IF(trx_status IN ('INITIATE', 'FAILED'), 0, ROUND(((fee_percentage/100) * trx_amount), 0)) AS exec_fee,
		IF(trx_status IN ('INITIATE', 'FAILED'), 0, ROUND(((fee_percentage/100) * trx_amount), 0) + trx_amount) AS exec_total_amount
	FROM (
		SELECT
			d.uuid AS id, IFNULL(at.uuid, '') AS trx_id, d.reference_id AS client_reference_id,
			IFNULL(d.metadata->>'$.cardFundedDetail.vendorId', '') AS vendor_id,
			IFNULL(d.metadata->>'$.cardFundedDetail.vendorName', '') AS vendor_name,
			d.beneficiary_bank_code AS bank_code, d.beneficiary_bank_name AS bank_name,
			d.beneficiary_account_no AS account_number, d.beneficiary_account_name AS account_name, IFNULL(d.remark, '') AS remarks,
			IFNULL(at.debit, IFNULL((SELECT SUM(amount) FROM payments WHERE merchant_id = d.merchant_id AND reference_id = d.uuid AND status = 'PAID'), d.amount)) AS trx_amount, 
			IFNULL(at.status, CASE WHEN p.status = 'PAID' THEN 'SCHEDULED' WHEN p.status IN ('CANCELLED', 'EXPIRED') THEN 'FAILED' ELSE 'INITIATE' END) AS trx_status, 
			at.reason_type AS trx_reason_type, at.reason_description AS trx_reason_desc, d.created_at, d.approved_at, 
			IFNULL(at.created_at, STR_TO_DATE(ap.additional_info->>'$.settlementDetail.estimateSettlementAt', '%Y-%m-%dT%H:%i:%sZ')) AS scheduled_at,
			at.created_at AS trx_created_at, CAST(at.updated_at AS DATETIME) AS trx_updated_at,
			d.merchant_id, d.amount AS init_amount, d.fee AS init_fee, d.total_amount AS init_total_amount,
			IFNULL(d.metadata->>'$.feeDetail.percentage', 0) AS fee_percentage, p.status AS first_payment_status
		FROM disbursements d
		LEFT JOIN account_transactions at ON at.reference_id = d.uuid AND (at.created_at BETWEEN ? AND DATE_ADD(?, INTERVAL 30 DAY))
		LEFT JOIN payments p ON p.merchant_id = d.merchant_id AND p.reference_id = d.uuid AND p.metadata->>'$.cardFundedPayout.sequence' = 1
		LEFT JOIN account_transactions ap ON ap.reference_id = p.uuid AND ap.type = 'PAYMENT'
		WHERE
			d.merchant_id = ?
			AND (d.created_at BETWEEN ? AND ?) 
			AND d.type = 'CARD_FUNDED_PAYOUT' AND d.status = 'APPROVED'
	) foo`

	conditions := []string{}
	args := []any{request.StartDate, request.EndDate, request.MerchantID, request.StartDate, request.EndDate}
	if request.TrxStatus != "" {
		args = append(args, request.TrxStatus)
		conditions = append(conditions, "foo.trx_status = ?")
	}

	if request.TrxReasonType != "" {
		args = append(args, request.TrxReasonType)
		if request.TrxReasonType == constant.ReasonTypeOtherReason {
			conditions = append(conditions, "(foo.trx_reason_type IS NULL OR foo.trx_reason_type = ?)")
		} else {
			conditions = append(conditions, "foo.trx_reason_type = ?")
		}
	}

	if len(conditions) > 0 {
		rawQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	rawQuery += " ORDER BY foo.created_at DESC"

	result := []cardFundedPayoutModel.GetPayoutTransactionListResponse{}
	if err := r.db.SelectContext(ctx, &result, rawQuery, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return result, nil
}
