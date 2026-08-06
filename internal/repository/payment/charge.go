package paymentRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PaymentRepository) GetChargeList(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetChargeList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	if request.Page < 1 {
		request.Page = 1
	}
	if request.PerPage < 1 {
		request.PerPage = 10
	}

	// Resolve sort column and direction
	sortBy := "att.created_at"
	switch request.SortBy {
	case "paymentDate":
		sortBy = "att.updated_at"
	}
	sortOrder := "DESC"
	if request.Sort != "" {
		sortOrder = request.Sort
	}

	// Build separate SELECT and COUNT queries
	fromClause := `
		FROM
			account_transactions att
			JOIN payments p
			ON p.uuid = att.reference_id AND p.metadata->"$.isUnifiedPaymentV2" = TRUE
			JOIN merchants m
			ON p.merchant_id = m.uuid`

	selectStatement := `SELECT
			att.uuid,
			att.merchant_id,
			m.name as merchant_name,
			p.uuid as 'payment_session_id',
			p.reference_id as 'payment_session_reference_id',
			CASE
				WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
					AND NULLIF(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','') IS NOT NULL
				THEN p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value'
				ELSE COALESCE(CAST(att.credit AS CHAR), '0')
			END AS 'amount.value',
			COALESCE(att.currency, "") as 'amount.currency',
			att.additional_info,
			COALESCE(att.additional_info->>'$.chargeStatus', "") as 'status',
			COALESCE(att.additional_info->>'$.statementDescriptor', "") as 'statement_descriptor',
			COALESCE(att.additional_info->>'$.methodDetail.qr', p.metadata->>'$.methodDetail.qr') as 'qr',
			COALESCE(att.additional_info->>'$.methodDetail.virtualAccount', p.metadata->>'$.methodDetail.virtualAccount') as 'virtual_account',
			att.additional_info->>'$.methodDetail.card' as 'card',
			att.additional_info->>'$.methodDetail.ewallet' as 'ewallet',
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN TRUE ELSE FALSE END as 'is_captured',
			CASE
				WHEN att.additional_info->>'$.chargeStatus' IN ('SUCCESS', 'WAITING_FOR_CAPTURE')
				THEN JSON_OBJECT('value', p.amount, 'currency', att.currency)
				ELSE NULL
			END AS authorized_amount,
			CASE
				WHEN att.additional_info->>'$.chargeStatus' IN ('SUCCESS', 'WAITING_FOR_CAPTURE')
				THEN JSON_OBJECT(
					'value',
					CASE
						WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
							AND NULLIF(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','') IS NOT NULL
						THEN CAST(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value' AS DOUBLE)
						ELSE att.credit
					END,
					'currency',
					att.currency
				)
				ELSE NULL
			END AS captured_amount,
			
			att.created_at as 'created_at',
			att.updated_at as 'updated_at',
			p.expired_at as expired_at,
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN att.transaction_timestamp ELSE NULL END as 'transaction_timestamp',
			CASE
				WHEN p.metadata->>'$.paymentMethod.type' = 'CARD' AND p.metadata->>'$.paymentMethodOptions.card.captureMethod' = 'MANUAL' THEN
					(
						SELECT JSON_ARRAYAGG(
							JSON_OBJECT(
								'captureId', pc.id,
								'status', pc.status,
								'currency', pc.currency,
								'capturedAmount', pc.amount,
								'createdAt', DATE_FORMAT(pc.created_at, '%Y-%m-%dT%H:%i:%sZ')
							)
						)
						FROM payment_captures pc
						WHERE pc.payment_id = p.uuid
					)
				ELSE NULL
			END AS capture_histories,
			IFNULL(att.settlement_status,'') as settlement_status`

	queryBuilder := util.QueryBuilder{
		SelectQuery: selectStatement + fromClause,
		CountQuery:  "SELECT COUNT(att.uuid)" + fromClause,
	}

	// Build filter
	conditions, args := buildWhereClauseForChargeLists(request)
	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	// Sort config
	sortConfig := util.SortConfig{
		DefaultSort: fmt.Sprintf("%s %s", sortBy, sortOrder),
	}

	// Destination slice
	data := make([]*unifiedPaymentModel.ChargeResponse, 0)

	// Data transformer — handles FDS extraction and response cleanup
	dataTransformer := func(result any) any {
		charges := *result.(*[]*unifiedPaymentModel.ChargeResponse)
		for _, respData := range charges {
			if respData.AdditionalInfo.Valid {
				var additionalInfo map[string]any
				if err := json.Unmarshal(
					respData.AdditionalInfo.JSONText, &additionalInfo,
				); err == nil {
					if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
						if fdsBytes, err := json.Marshal(fdsData); err == nil {
							var fdsAssessment fdsCommonModel.FdsRiskAssessment
							if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
								respData.FdsRiskAssessment = &fdsAssessment
							}
						}
					}
				}
			}
			respData.SetCaptureHistoriesFromJSON()
			respData.SetFailureDetail()
			respData.RemoveUnusedResponse()
		}
		return charges
	}

	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      r.appConfig.InitialPageWindow,
	}
	paginationUtil := util.NewPaginationUtility(r.db, r.logger, paginationConfig)

	return paginationUtil.GetPaginatedList(
		ctx, queryBuilder, filterResult, sortConfig,
		int64(request.Page), int64(request.PerPage),
		&data, dataTransformer,
	)
}

func (r *PaymentRepository) GetChargeByID(ctx context.Context, chargeID string) (*unifiedPaymentModel.ChargeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetChargeByID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		result = &unifiedPaymentModel.ChargeResponse{}
	)

	query := `
		SELECT
			att.uuid,
			att.merchant_id,
			p.uuid as 'payment_session_id',
			p.reference_id as 'payment_session_reference_id',
			CASE
				WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
					AND NULLIF(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','') IS NOT NULL
				THEN p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value'
				ELSE COALESCE(CAST(att.credit AS CHAR), '0')
			END AS 'amount.value',
			COALESCE(att.currency, "") as 'amount.currency',
			att.additional_info,
			COALESCE(att.additional_info->>'$.chargeStatus', "") as 'status',
			COALESCE(att.additional_info->>'$.statementDescriptor', "") as 'statement_descriptor',
			COALESCE(att.additional_info->>'$.methodDetail.qr', p.metadata->>'$.methodDetail.qr') as 'qr',
			COALESCE(att.additional_info->>'$.methodDetail.virtualAccount', p.metadata->>'$.methodDetail.virtualAccount') as 'virtual_account',
			att.additional_info->>'$.methodDetail.card' as 'card',
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN TRUE ELSE FALSE END as 'is_captured',
			CASE
				WHEN att.additional_info->>'$.chargeStatus' IN ('SUCCESS', 'WAITING_FOR_CAPTURE')
				THEN JSON_OBJECT(
					'value',
					CASE
						WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
							AND NULLIF(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','') IS NOT NULL
						THEN CAST(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value' AS DOUBLE)
						ELSE p.total_amount
					END,
					'currency',
					att.currency
				)
				ELSE NULL
			END AS authorized_amount,
			CASE
				WHEN att.additional_info->>'$.chargeStatus' IN ('SUCCESS', 'WAITING_FOR_CAPTURE')
				THEN JSON_OBJECT(
					'value',
					CASE
						WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
							AND NULLIF(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value','') IS NOT NULL
						THEN CAST(p.metadata->>'$.autoSplitPayment.summary.totalSuccessfulChargeAmount.value' AS DOUBLE)
						ELSE att.credit
					END,
					'currency',
					att.currency
				)
				ELSE NULL
			END AS captured_amount,
			att.created_at as 'created_at',
			att.updated_at as 'updated_at',
			p.expired_at as 'expired_at',
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN att.transaction_timestamp ELSE NULL END as 'transaction_timestamp',
			CASE
				WHEN p.metadata->>'$.paymentMethod.type' = 'CARD' AND p.metadata->>'$.paymentMethodOptions.card.captureMethod' = 'MANUAL' THEN
					(
						SELECT JSON_ARRAYAGG(
							JSON_OBJECT(
								'captureId', pc.id,
								'status', pc.status,
								'currency', pc.currency,
								'capturedAmount', pc.amount,
								'createdAt', DATE_FORMAT(pc.created_at, '%Y-%m-%dT%H:%i:%sZ')
							)
						)
						FROM payment_captures pc
						WHERE pc.payment_id = p.uuid
					)
				ELSE NULL
			END AS capture_histories
		FROM
			account_transactions att
			JOIN payments p
			ON p.uuid = att.reference_id AND p.metadata->"$.isUnifiedPaymentV2" = TRUE
			WHERE att.type = ? AND att.uuid = ?`

	err := r.db.GetContext(ctx, result, query, constant.TypePayment, chargeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		r.logger.Error(ctx, "error when get charge id", logger.Error(err))
		return nil, err
	}

	// Extract FDS risk assessment from additional_info
	if result.AdditionalInfo.Valid {
		var additionalInfo map[string]interface{}
		if err := json.Unmarshal(result.AdditionalInfo.JSONText, &additionalInfo); err == nil {
			if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
				// Convert the fdsData to JSON and then unmarshal to FdsRiskAssessment
				if fdsBytes, err := json.Marshal(fdsData); err == nil {
					var fdsAssessment fdsCommonModel.FdsRiskAssessment
					if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
						result.FdsRiskAssessment = &fdsAssessment
					}
				}
			}
		}
	}

	result.SetCaptureHistoriesFromJSON()
	result.SetFailureDetail()
	result.RemoveUnusedResponse()

	return result, nil
}

// GetCharges retrieves a list of charge transactions based on the provided filter criteria. this is simple version of GetChargeList
// It supports sorting and filtering by various fields as specified in the FilterChargeRequest.
// The function joins account transactions with payments and merchants, and returns a slice of ChargeResponse.
// It returns an error if the database query fails.
// Note: This function does not apply pagination and returns all matching records, typically used for exports.
func (r *PaymentRepository) GetCharges(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) ([]unifiedPaymentModel.ChargeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetCharges")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		result         = make([]unifiedPaymentModel.ChargeResponse, 0)
		whereStatement string
		sortBy         string = "att.created_at"
		sortOrder      string = "DESC"
	)

	if request.Sort != "" {
		sortOrder = request.Sort
	}

	switch request.SortBy {
	case "createdAt":
		sortBy = "att.created_at"
	case "paymentDate":
		sortBy = "att.updated_at" // sort by the account_transaction updated_at
	}

	queryTemplate := `
		SELECT
			%s
		FROM
			account_transactions att
			JOIN payments p
			ON p.uuid = att.reference_id AND p.metadata->"$.isUnifiedPaymentV2" = TRUE
			JOIN merchants m
			ON p.merchant_id = m.uuid`

	selectStatement := `att.uuid,
			att.merchant_id,
			m.name as merchant_name,
			p.uuid as 'payment_session_id',
			p.reference_id as 'payment_session_reference_id',
			COALESCE(att.credit, 0) as 'amount.value',
			COALESCE(att.currency, "") as 'amount.currency',
			att.additional_info,
			COALESCE(att.additional_info->>'$.chargeStatus', "") as 'status',
			COALESCE(att.additional_info->>'$.failureCode', "") as 'failure_code',
			COALESCE(att.additional_info->>'$.statementDescriptor', "") as 'statement_descriptor',
			COALESCE(att.additional_info->>'$.methodDetail.qr', p.metadata->>'$.methodDetail.qr') as 'qr',
			COALESCE(att.additional_info->>'$.methodDetail.virtualAccount', p.metadata->>'$.methodDetail.virtualAccount') as 'virtual_account',
			att.additional_info->>'$.methodDetail.card' as 'card',
			att.additional_info->>'$.methodDetail.ewallet' as 'ewallet',
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN TRUE ELSE FALSE END as 'is_captured',
			CASE
				WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
				THEN JSON_OBJECT('value', att.credit, 'currency', att.currency)
				ELSE NULL
			END AS authorized_amount,
			CASE
				WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS'
				THEN JSON_OBJECT('value', att.credit, 'currency', att.currency)
				ELSE NULL
			END AS captured_amount,
			att.created_at as 'created_at',
			att.updated_at as 'updated_at',
			p.expired_at as expired_at,
			CASE WHEN att.additional_info->>'$.chargeStatus' = 'SUCCESS' THEN att.transaction_timestamp ELSE NULL END as 'transaction_timestamp'`
	whereClause, args := buildWhereClauseForChargeLists(request)
	if len(whereClause) > 0 {
		whereStatement = " WHERE " + strings.Join(whereClause, " AND ")
	}

	sortStatement := fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	query := fmt.Sprintf(queryTemplate, selectStatement)
	query = query + whereStatement + sortStatement

	err := r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when get charges", logger.Error(err))
		return nil, err
	}

	for i := range result {
		result[i].SetFailureDetail()
		result[i].RemoveUnusedResponse()
	}

	return result, nil
}

func buildWhereClauseForChargeLists(request *unifiedPaymentModel.FilterChargeRequest) (whereClause []string, args []any) {
	// Filter by merchant and 'PAYMENT' type to ensure data accuracy and query optimization.
	whereClause = append(whereClause, "att.merchant_id = ? AND att.type = ?")
	args = append(args, request.MerchantID, constant.TypePayment)

	if len(request.PaymentTypes) == 0 {
		request.PaymentTypes = []string{"", constant.PaymentTypeMultiple, constant.PaymentTypeSingle}
	}

	paymentTypePlaceholders := make([]string, len(request.PaymentTypes))
	for i, v := range request.PaymentTypes {
		paymentTypePlaceholders[i] = "?"
		args = append(args, v)
	}
	whereClause = append(whereClause, fmt.Sprintf("p.type IN (%s)", strings.Join(paymentTypePlaceholders, ", ")))

	if request.UUID != "" {
		whereClause = append(whereClause, "att.uuid = ?")
		args = append(args, request.UUID)
	}

	if request.PaymentSessionID != "" {
		whereClause = append(whereClause, "p.uuid = ?")
		args = append(args, request.PaymentSessionID)
	}

	if request.Status != "" {
		statuses := strings.Split(request.Status, ",")
		statusQuery, statusArgs, _ := sqlx.In("att.additional_info->>'$.chargeStatus' IN (?)", statuses)
		whereClause = append(whereClause, statusQuery)
		args = append(args, statusArgs...)
	}

	if !request.StartCreatedAt.IsZero() || !request.EndCreatedAt.IsZero() {
		dateRange := []string{}
		if !request.StartCreatedAt.IsZero() {
			dateRange = append(dateRange, "att.created_at >= ?")
			args = append(args, request.StartCreatedAt)
		}
		if !request.EndCreatedAt.IsZero() {
			dateRange = append(dateRange, "att.created_at <= ?")
			args = append(args, request.EndCreatedAt)
		}

		whereClause = append(whereClause, fmt.Sprintf("(%s)", strings.Join(dateRange, " AND ")))
	}

	if !request.StartPaymentDate.IsZero() || !request.EndPaymentDate.IsZero() {
		dateRange := []string{}
		if !request.StartPaymentDate.IsZero() {
			dateRange = append(dateRange, "att.updated_at >= ?")
			args = append(args, request.StartPaymentDate)
		}
		if !request.EndPaymentDate.IsZero() {
			dateRange = append(dateRange, "att.updated_at <= ?")
			args = append(args, request.EndPaymentDate)
		}

		whereClause = append(whereClause, fmt.Sprintf("(%s)", strings.Join(dateRange, " AND ")))
	}

	if request.ClientReferenceID != "" {
		whereClause = append(whereClause, "(p.reference_id LIKE ? OR p.uuid LIKE ? OR att.uuid LIKE ?)")
		args = append(args, "%"+request.ClientReferenceID+"%", "%"+request.ClientReferenceID+"%", "%"+request.ClientReferenceID+"%")
	}

	return whereClause, args
}
