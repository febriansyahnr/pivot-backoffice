package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (r *AccountTransactionRepository) buildConditionForTransactionHistory(filter *orchestratorModel.TransactionHistoryFilterRequest) (conditions []string, args []interface{}) {
	if filter.MerchantID != "" {
		conditions = append(conditions, "(t.merchant_id = ?)")
		args = append(args, filter.MerchantID)
	}

	// created_at filter for query optimization
	if !filter.CreatedAtStartDate.IsZero() && !filter.CreatedAtEndDate.IsZero() {
		args = append(args, filter.CreatedAtStartDate, filter.CreatedAtEndDate)
		conditions = append(conditions, "(t.created_at BETWEEN ? AND ?)")
	}

	if !filter.StartSettlementDate.IsZero() && !filter.EndSettlementDate.IsZero() {
		// Add updated_at filter first for index optimization
		args = append(args, filter.StartSettlementDate, filter.EndSettlementDate)
		conditions = append(conditions, "(t.updated_at BETWEEN ? AND ?)")

		args = append(args, filter.StartSettlementDate, filter.EndSettlementDate, filter.StartSettlementDate, filter.EndSettlementDate, filter.StartSettlementDate, filter.EndSettlementDate)
		conditions = append(conditions, `
			(
				(t.settlement_status = 'SUCCESS'
					AND t.settlement_at IS NOT NULL
					AND t.settlement_at != '0001-01-01T00:00:00Z'
					AND CAST(t.settlement_at AS DATETIME) BETWEEN ? AND ?)
				OR
				(t.settlement_status = 'PENDING'
					AND t.additional_info->>'$.settlementDetail.type' LIKE 'T+%'
					AND DATE_ADD(t.updated_at, INTERVAL CAST(SUBSTRING(t.additional_info->>'$.settlementDetail.type', 3) AS SIGNED) DAY) BETWEEN ? AND ?)
				OR
				((t.settlement_status IS NULL OR t.settlement_status = 'PENDING')
					AND (t.additional_info->>'$.deductionType' = 'DIRECT' OR t.settlement_at IS NULL OR t.settlement_at = '0001-01-01T00:00:00Z' OR t.additional_info->>'$.settlementDetail.type' NOT LIKE 'T+%')
					AND t.updated_at BETWEEN ? AND ?)
			)
		`)
	}

	// Multi-select for balance types
	if len(filter.BalanceTypes) > 0 {
		var orConds []string
		for _, balanceType := range filter.BalanceTypes {
			orConds = append(orConds, "a.name = ?")
			args = append(args, balanceType)
		}
		conditions = append(conditions, "("+strings.Join(orConds, " OR ")+")")
	}

	// Multi-select for transaction types, with special-case logic
	if len(filter.TrxTypes) > 0 {
		var orConds []string
		var orArgs []interface{}
		for _, trxType := range filter.TrxTypes {
			switch trxType {
			case "DISBURSEMENT":
				orConds = append(orConds, "(t.type = ? AND d.bulk_id IS NULL AND t.channel != 'XB')")
				orArgs = append(orArgs, trxType)
			case "BULK_DISBURSEMENT":
				orConds = append(orConds, "(t.type = ? AND d.bulk_id IS NOT NULL)")
				orArgs = append(orArgs, constant.TypeDisbursement)
			case "INTERNATIONAL_PAYOUT":
				orConds = append(orConds, "(t.type = ? AND t.channel = 'XB')")
				orArgs = append(orArgs, constant.TypeDisbursement)
			case "MANUAL_TOP_UP":
				orConds = append(orConds, "(t.type = ? AND t.channel = 'MANUAL_TRANSFER')")
				orArgs = append(orArgs, constant.TypeManualAdjust)
			case "BALANCE_ADJUSTMENT":
				orConds = append(orConds, "(t.type = ? AND t.channel = 'BALANCE_ADJUSTMENT')")
				orArgs = append(orArgs, constant.TypeManualAdjust)
			case "VA_TOP_UP":
				orConds = append(orConds, "(t.type IN (?, ?) AND t.channel = 'VIRTUAL_ACCOUNT')")
				orArgs = append(orArgs, constant.TypeGeneralTopUp, constant.TypeTopUp)
			case "CUSTOMER_TOP_UP":
				orConds = append(orConds, "(t.type = ? AND t.channel = 'MANUAL_TRANSFER' and t.reference = 'WALLET')")
				orArgs = append(orArgs, constant.TypeGeneralTopUp)
			case "PAYMENT_WITHDRAWAL", "DISBURSEMENT_WITHDRAWAL", "WALLET_WITHDRAWAL":
				orConds = append(orConds, "(t.type = ? AND t.reference = ?)")
				orArgs = append(orArgs, constant.TypeWithdrawal, strings.TrimSuffix(trxType, "_WITHDRAWAL"))
			case "VA_PAYMENT":
				orConds = append(orConds, "(t.type = ? AND t.channel = ?)")
				orArgs = append(orArgs, constant.TypePayment, constant.ChannelVirtualAccount)
			case "QRIS_PAYMENT":
				orConds = append(orConds, "(t.type = ? AND t.channel = ?)")
				orArgs = append(orArgs, constant.TypePayment, constant.ChannelQris)
			case "CARD_PAYMENT":
				orConds = append(orConds, "(t.type = ? AND t.channel = ?)")
				orArgs = append(orArgs, constant.TypePayment, constant.ChannelCard)
			case "WALLET_PAYMENT":
				orConds = append(orConds, "(t.type = ? AND t.channel = ?)")
				orArgs = append(orArgs, constant.TypePayment, constant.ChannelEwallet)
			default:
				if strings.HasSuffix(trxType, "_FEE") {
					useCaseType := strings.TrimSuffix(trxType, "_FEE")
					orConds = append(orConds, "(t.type = ? AND t.additional_info->>'$.type' = ?)")
					orArgs = append(orArgs, constant.TypeFee, useCaseType)
				} else {
					orConds = append(orConds, "(t.type = ?)")
					orArgs = append(orArgs, trxType)
				}
			}
		}
		if len(orConds) > 0 {
			conditions = append(conditions, "("+strings.Join(orConds, " OR ")+")")
			args = append(args, orArgs...)
		}
	}

	if filter.Status != "" {
		switch strings.ToUpper(filter.Status) {
		case "SUCCESS":
			conditions = append(conditions, "(settlement_status IS NULL OR settlement_status = 'SUCCESS')")

		case "PENDING":
			conditions = append(conditions, "(settlement_status = 'PENDING')")
		}
	}

	if filter.TrxID != "" {
		conditions = append(conditions, "t.reference_id = ?")
		args = append(args, filter.TrxID)
	}

	if filter.TransactionId != "" {
		// Improved condition to handle multiple payments with the same reference_id
		conditions = append(conditions, `(
			t.uuid = ? OR t.reference_id = ? OR t.merchant_reference_id = ? OR d.reference_id = ? OR p.reference_id = ? OR trf.reference_id = ?
		)`)
		args = append(args, filter.TransactionId, filter.TransactionId, filter.TransactionId, filter.TransactionId, filter.TransactionId, filter.TransactionId)
	}

	if filter.MerchantReferenceID != "" {
		conditions = append(conditions, "(t.merchant_reference_id = ? OR d.reference_id = ? OR p.reference_id = ? OR trf.reference_id = ?)")
		args = append(args, filter.MerchantReferenceID, filter.MerchantReferenceID, filter.MerchantReferenceID, filter.MerchantReferenceID)
	}

	if filter.SettlementModel != "" {
		if filter.SettlementModel == constant.PaymentMethodChannelTypeAggregator {
			conditions = append(conditions, "(t.settlement_model = ? OR t.settlement_model IS NULL)")
		} else {
			conditions = append(conditions, "t.settlement_model = ?")
		}
		args = append(args, filter.SettlementModel)
	}

	// Exclude Card Funded Payout Balance
	conditions = append(conditions, "a.name IN ('DISBURSEMENT', 'PAYMENT', 'WALLET', 'VIRTUAL_TERMINAL')")

	/*
	 * Exceptions to payment fees, disbursement fees, transfer fees, sub payments and wallet transaction fees deducted from merchants.
	 * To Do (Added Condition): OR t.additional_info->>'$.type' = 'PLATFORM_TRANSACTION' OR (t.debit > 0 AND t.additional_info->>'$.referenceType' = 'TOP_UP')
	 */
	conditions = append(conditions, `
		NOT (t.type = 'FEE' AND ((p.uuid IS NOT NULL AND p.merchant_id = t.merchant_id) OR (d.uuid IS NOT NULL AND d.merchant_id = t.merchant_id) OR (mtr.uuid IS NOT NULL)))
		AND (t.status = 'SUCCESS' OR (t.status = 'PENDING' AND t.type = 'FEE'))
		AND NOT (t.remarks LIKE '%Disbursement Fee Transfer%' AND t.debit > 0)
		AND t.reference != 'SUB_PAYMENT'
	`)
	return
}

func (r *AccountTransactionRepository) GetList(
	ctx context.Context,
	filter *orchestratorModel.TransactionHistoryFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	// Initialize pagination utility
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
			t.uuid, t.reference_id, t.account_id, t.currency, t.settlement_model,
			IFNULL(t.additional_info->>'$.subPaymentSummary.totalCreditAmount', t.credit) AS credit, 
			t.debit,
			CASE 
    			WHEN t.type = 'PAYMENT' AND t.additional_info->>'$.feeDetail.finalAmount' IS NOT NULL THEN t.additional_info->>'$.feeDetail.finalAmount'
    			WHEN t.type = 'PAYMENT' AND p.metadata->>'$.feeDetail.finalAmount' IS NOT NULL THEN p.metadata->>'$.feeDetail.finalAmount'
    			WHEN t.type = 'PAYMENT' AND t.additional_info->>'$.subPaymentSummary.totalFeeAmount' IS NOT NULL THEN t.additional_info->>'$.subPaymentSummary.totalFeeAmount'
    			WHEN t.type = 'DISBURSEMENT' THEN d.fee
				WHEN t.type = 'DISBURSEMENT_TOP_UP' THEN IFNULL(t.additional_info->>'$.feeDetail.finalAmount', IFNULL(p.metadata->>'$.feeDetail.finalAmount', 0))
				WHEN (t.type = 'TOP_UP' AND t.reference = 'WALLET') OR (t.type = 'TRANSFER' AND t.reference = 'PLATFORM') THEN 
					IFNULL((SELECT debit FROM account_transactions WHERE reference_id = t.reference_id AND merchant_id = t.merchant_id AND type = 'FEE' AND IF(t.type = 'TRANSFER', TRUE, (additional_info->>'$.referenceType' = 'TOP_UP')) AND debit > 0 LIMIT 1), 0)
    			ELSE 0
    		END AS fee,
			CASE 
				WHEN t.type = 'DISBURSEMENT' AND d.bulk_id IS NOT NULL THEN 'BULK_DISBURSEMENT'
				WHEN t.type = 'FEE' AND mtr.uuid IS NOT NULL THEN 'DISBURSEMENT_TOP_UP_FEE'
				WHEN t.type = 'FEE' AND t.reference != 'WALLET' THEN IFNULL(CONCAT(t.additional_info->>'$.type', '_FEE'), 'FEE')
				WHEN t.type = 'WITHDRAWAL' THEN CONCAT(TRIM(t.reference), '_', TRIM(t.type))
				ELSE t.type
			END AS type, IF(t.reference = 'WALLET', IF(t.channel != '', t.channel, IFNULL(t.additional_info->>'$.referenceType', '')), t.channel) AS channel,
			IFNULL(t.settlement_status, 'SUCCESS') AS status,
			t.transaction_timestamp, t.created_at, t.updated_at,
			CASE
				WHEN t.settlement_status = 'SUCCESS' 
					AND t.settlement_at != '0001-01-01T00:00:00Z' 
					AND t.settlement_at IS NOT NULL THEN
						CAST(t.settlement_at AS DATETIME)
				WHEN t.settlement_status = 'PENDING' 
					AND t.additional_info->>'$.settlementDetail.estimateSettlementAt' IS NOT NULL THEN
						STR_TO_DATE(t.additional_info->>'$.settlementDetail.estimateSettlementAt', '%Y-%m-%dT%H:%i:%sZ')
				WHEN t.settlement_status = 'PENDING' 
					AND t.additional_info->>'$.settlementDetail.type' LIKE 'T+%' THEN
						DATE_ADD(t.updated_at, INTERVAL CAST(SUBSTRING(t.additional_info->>'$.settlementDetail.type', 3) AS SIGNED) DAY)
				WHEN t.additional_info->>'$.deductionType' = 'DIRECT' 
					OR t.settlement_at IS NULL 
					OR t.settlement_at = '0001-01-01T00:00:00Z' THEN
						t.updated_at
				ELSE t.updated_at
			END AS settlement_at,
			IF(t.type NOT IN ('DISBURSEMENT', 'WITHDRAWAL'), '-', IFNULL(d.beneficiary_account_name, w.beneficiary_account_name)) AS beneficiary_account_name,
			CASE 
				WHEN t.type IN ('FEE','DISBURSEMENT_TOP_UP','MANUAL_ADJUSTMENT','MERCHANT_TOP_UP') THEN 'System'
				WHEN t.type IN ('PAYMENT','CASHBACK','MERCHANT_PAYMENT') THEN IFNULL(m.name, '')
				WHEN t.type = 'DISBURSEMENT' THEN IFNULL(d_users.name, m.name)
				WHEN t.type = 'WITHDRAWAL' THEN IFNULL(w_users.name, 'System')
				WHEN t.type = 'TRANSFER' AND (t.remarks LIKE '%Fee Transfer%') THEN 'System'
				WHEN t.type = 'TRANSFER' THEN IFNULL(m.name, 'System')
				WHEN t.type = 'TOP_UP' THEN IFNULL(m.name, 'System')
			ELSE '-' END AS created_by,
			CASE 
        		WHEN a.name = 'DISBURSEMENT' THEN 'Payout Balance'
        		WHEN a.name = 'PAYMENT' THEN 'Payment Balance'
				WHEN a.name = 'WALLET' THEN 'Wallet Balance'
				WHEN a.name = 'VIRTUAL_TERMINAL' THEN 'Virtual Terminal Balance'
        		ELSE '-' 
    		END AS balance_type,
			CASE
        		WHEN t.type = 'DISBURSEMENT' THEN IFNULL(d.reference_id, '-')
        		WHEN t.type = 'PAYMENT' THEN IFNULL(p.reference_id, '-')
				WHEN t.type = 'TRANSFER' THEN IFNULL(trf.reference_id, '-')
        		WHEN t.type = 'FEE' AND t.reference = 'DISBURSEMENT' THEN IFNULL(d.reference_id, '-')
        		WHEN t.type = 'FEE' AND t.reference = 'PAYMENT' THEN IFNULL(p.reference_id, '-')
        		WHEN t.type = 'MANUAL_ADJUSTMENT' AND t.channel = 'BALANCE_ADJUSTMENT' THEN IFNULL(mah.reference_id, '-')
        		WHEN t.type = 'REFUND' THEN IFNULL(ref_p.reference_id, '-')
        		ELSE IFNULL(t.merchant_reference_id, '-')
    		END AS merchant_reference_id
        FROM account_transactions t
		JOIN accounts a ON a.uuid = t.account_id AND a.user_type = 'MERCHANT'
        LEFT JOIN disbursements d ON d.uuid = t.reference_id
		LEFT JOIN payments p ON p.uuid = t.reference_id
		LEFT JOIN withdrawals w ON w.id = t.reference_id
		LEFT JOIN transfers trf ON trf.uuid = t.reference_id
		LEFT JOIN manual_adjustment_histories mah ON mah.uuid = t.reference_id
		LEFT JOIN refunds rf ON rf.uuid = t.reference_id AND t.type = 'REFUND'
		LEFT JOIN payments ref_p ON ref_p.uuid = rf.payment_id
		LEFT JOIN users w_users ON w_users.uuid = w.created_by
  		LEFT JOIN users d_users ON d_users.uuid = d.created_by
		LEFT JOIN merchant_top_up_references mtr ON t.reference_id = mtr.uuid
		LEFT JOIN merchants m ON m.uuid = IFNULL(trf.merchant_id, t.merchant_id)`,
		CountQuery: `SELECT 
				COUNT(t.uuid) as totalItems
		FROM account_transactions t
		JOIN accounts a ON a.uuid = t.account_id AND a.user_type = 'MERCHANT'
        LEFT JOIN disbursements d ON d.uuid = t.reference_id
		LEFT JOIN payments p ON p.uuid = t.reference_id
		LEFT JOIN withdrawals w ON w.id = t.reference_id
		LEFT JOIN transfers trf ON trf.uuid = t.reference_id`,
	}

	// Build filter conditions
	conditions, args := r.buildConditionForTransactionHistory(filter)
	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	// Build sort configuration
	sortConfig := util.SortConfig{
		DefaultSort: filter.FilteredSortQuery,
		SortBy:      "",
		Sort:        "",
	}

	// Data destination
	data := make([]*orchestratorModel.AccountTransactionWithUseCase, 0)

	// Data transformer
	featureName, _ := ctx.Value(constant.CtxFeatureName).(string)
	dataTransformer := func(dest interface{}) interface{} {
		typedData := dest.(*[]*orchestratorModel.AccountTransactionWithUseCase)
		response := make([]any, 0, len(*typedData))
		for _, item := range *typedData {
			if featureName == constant.FeatureBalanceHistoryOpenApi {
				response = append(response, orchestratorModel.ToTransactionHistoryOpenApiResponse(item))
			} else {
				response = append(response, orchestratorModel.ToTransactionHistoryResponse(item))
			}
		}
		return response
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

func (r *AccountTransactionRepository) GetListTransactionHistories(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest) (result []orchestratorModel.TransactionHistory, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetListTransactionHistories")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "multiples-tb-transaction-histories-excel")

	rawQuery := `SELECT
			t.created_at, t.updated_at,
			CASE
				WHEN t.settlement_status = 'SUCCESS' 
					AND t.settlement_at != '0001-01-01T00:00:00Z' 
					AND t.settlement_at IS NOT NULL THEN
						CAST(t.settlement_at AS DATETIME)
				WHEN t.settlement_status = 'PENDING' 
					AND t.additional_info->>'$.settlementDetail.estimateSettlementAt' IS NOT NULL THEN
						STR_TO_DATE(t.additional_info->>'$.settlementDetail.estimateSettlementAt', '%Y-%m-%dT%H:%i:%sZ')
			    WHEN t.settlement_status = 'PENDING' 
					AND t.additional_info->>'$.settlementDetail.type' LIKE 'T+%' THEN
			        	DATE_ADD(t.updated_at, INTERVAL CAST(SUBSTRING(t.additional_info->>'$.settlementDetail.type', 3) AS SIGNED) DAY)
			    WHEN t.additional_info->>'$.deductionType' = 'DIRECT' 
					OR t.settlement_at IS NULL 
					OR t.settlement_at = '0001-01-01T00:00:00Z' THEN
			        	t.updated_at
			    ELSE CAST(t.settlement_at AS DATETIME)
			END AS settlement_at,
			CASE
				WHEN t.type = 'DISBURSEMENT' AND d.bulk_id IS NOT NULL THEN 'BULK_DISBURSEMENT'  
				WHEN t.type = 'FEE' AND mtr.uuid IS NOT NULL THEN 'DISBURSEMENT_TOP_UP_FEE'
				WHEN t.type = 'FEE' AND t.reference != 'WALLET' THEN IFNULL(CONCAT(t.additional_info->>'$.type', '_FEE'), 'FEE')
				ELSE t.type
			END AS type, IF(t.reference = 'WALLET', IF(t.channel != '', t.channel, IFNULL(t.additional_info->>'$.referenceType', '')), t.channel) AS channel, 
			CASE 
				WHEN t.type IN ('FEE','DISBURSEMENT_TOP_UP','MANUAL_ADJUSTMENT','MERCHANT_TOP_UP') THEN 'System'
				WHEN t.type IN ('PAYMENT','CASHBACK','MERCHANT_PAYMENT') THEN IFNULL(m.name, '')
				WHEN t.type = 'DISBURSEMENT' THEN IFNULL(users.name, m.name)
				WHEN t.type = 'WITHDRAWAL' THEN IFNULL(users.name, 'System')
				WHEN t.type = 'TRANSFER' AND (t.remarks LIKE '%Fee Transfer%') THEN 'System'
				WHEN t.type = 'TRANSFER' THEN IFNULL(m.name, 'System')
				WHEN t.type = 'TOP_UP' THEN IFNULL(users.name, 'System')
			ELSE '-' 
			END AS created_by, 
			t.uuid AS id, IFNULL(t.additional_info->>'$.linked_transaction_id', '') AS linked_id, IFNULL(d.reference_id, '') AS reference_id, 
			IFNULL(d.reference_id, '') AS reference_id, IF(t.credit > 0, t.credit, -1*t.debit) AS amount,
			CASE 
    			WHEN t.type = 'PAYMENT' THEN IFNULL(t.additional_info->>'$.feeDetail.finalAmount', IFNULL(p.metadata->>'$.feeDetail.finalAmount', 0))
    			WHEN t.type = 'DISBURSEMENT' THEN d.fee
				WHEN t.type = 'DISBURSEMENT_TOP_UP' THEN IFNULL(t.additional_info->>'$.feeDetail.finalAmount', IFNULL(p.metadata->>'$.feeDetail.finalAmount', 0))
				WHEN (t.type = 'TOP_UP' AND t.reference = 'WALLET') OR (t.type = 'TRANSFER' AND t.reference = 'PLATFORM') THEN 
					IFNULL((SELECT debit FROM account_transactions WHERE reference_id = t.reference_id AND merchant_id = t.merchant_id AND type = 'FEE' AND IF(t.type = 'TRANSFER', TRUE, (additional_info->>'$.referenceType' = 'TOP_UP')) AND debit > 0 LIMIT 1), 0)
    			ELSE 0
    		END AS fee,
			IFNULL(d.bank_reference_no, IF(t.type = 'MANUAL_ADJUSTMENT' AND t.channel = 'MANUAL_TRANSFER', adj.bank_reference_id, '')) AS bank_reference, 	
			IFNULL(t.settlement_status, 'SUCCESS') AS status,
			IFNULL(t.reason_type, '') AS reason_type, 
			IFNULL(t.reason_description, '') AS reason_description, 
			IF(t.type = 'FEE', '', IFNULL(t.remarks, '')) AS remarks, IFNULL(d.beneficiary_bank_name, '') AS beneficiary_bank_name, 
			IFNULL(d.beneficiary_account_no, '') AS beneficiary_account_no, IFNULL(d.beneficiary_account_name, '') AS beneficiary_account_name,
			IFNULL(d.status, '') AS approval_status, d.approved_at, 
			IF(NULLIF(d.approved_by, '') IS NULL AND d.status IS NOT NULL, 'AUTO', IFNULL(users.name, '')) AS approved_by,
			CASE 
        		WHEN a.name = 'DISBURSEMENT' THEN 'Payout Balance'
        		WHEN a.name = 'PAYMENT' THEN 'Payment Balance'
				WHEN a.name = 'WALLET' THEN 'Wallet Balance'
        		ELSE '-' 
    		END AS balance_type,
			CASE
				WHEN t.type = 'DISBURSEMENT' THEN IFNULL(d.reference_id, '-')
        		WHEN t.type = 'PAYMENT' THEN IFNULL(p.reference_id, '-')
				WHEN t.type = 'TRANSFER' THEN IFNULL(trf.reference_id, '-')
        		WHEN t.type = 'FEE' AND t.reference = 'DISBURSEMENT' THEN IFNULL(d.reference_id, '-')
        		WHEN t.type = 'FEE' AND t.reference = 'PAYMENT' THEN IFNULL(p.reference_id, '-')
        		WHEN t.type = 'REFUND' THEN IFNULL(ref_p.reference_id, '-')
        		ELSE IFNULL(t.merchant_reference_id, '-')
    		END AS merchant_reference_id
		FROM account_transactions t
		JOIN accounts a ON a.uuid = t.account_id AND a.user_type = 'MERCHANT'
		LEFT JOIN disbursements d ON d.uuid = t.reference_id
		LEFT JOIN payments p ON p.uuid = t.reference_id
		LEFT JOIN manual_adjustment_histories adj ON adj.uuid = t.reference_id AND t.type = 'MANUAL_ADJUSTMENT'
		LEFT JOIN withdrawals w ON w.id = t.reference_id
		LEFT JOIN transfers trf ON trf.uuid = t.reference_id
		LEFT JOIN refunds rf ON rf.uuid = t.reference_id AND t.type = 'REFUND'
		LEFT JOIN payments ref_p ON ref_p.uuid = rf.payment_id
		LEFT JOIN users ON (users.uuid = w.created_by) OR (users.uuid = d.approved_by)
		LEFT JOIN merchant_top_up_references mtr ON t.reference_id = mtr.uuid
		LEFT JOIN merchants m ON m.uuid = IFNULL(trf.merchant_id, t.merchant_id)`

	conditions, args := r.buildConditionForTransactionHistory(filter)

	if len(conditions) > 0 {
		rawQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	rawQuery += " ORDER BY " + filter.FilteredSortQuery

	if err = r.db.SelectContext(ctx, &result, rawQuery, args...); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
