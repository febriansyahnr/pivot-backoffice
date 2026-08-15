package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) GetAggregateTransactions(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (*orchestratorModel.AggregateResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetAggregateTransactions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var (
		aggregate            orchestratorModel.AggregateResponse
		calculatePendTrxExpr = "reference NOT IN ('PAYMENT', 'WALLET', 'VIRTUAL_TERMINAL') AND status = 'PENDING'"
	)

	query := fmt.Sprintf(`
			SELECT 
				COUNT(CASE WHEN credit > 0 THEN 1 END) AS count_of_credit,
				COUNT(CASE WHEN debit > 0 THEN 1 END) AS count_of_debit,
				IFNULL(SUM(credit), 0) AS sum_of_credit, 
				IFNULL(SUM(debit), 0) AS sum_of_debit,
				IFNULL(SUM(IF(%s, credit, 0)), 0) AS sum_of_pend_credit,
				IFNULL(SUM(IF(%s, debit, 0)), 0) AS sum_of_pend_debit 
			FROM account_transactions`, calculatePendTrxExpr, calculatePendTrxExpr,
	)

	query = query + r.constructTotalBalanceWhereClause(request)
	if err := r.db.GetContext(ctx, &aggregate, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when get aggregate transaction", logger.Error(err))
		return nil, err
	}
	return &aggregate, nil
}

func (r *AccountTransactionRepository) constructTotalBalanceWhereClause(request *orchestratorModel.GetAggregateRequest) string {
	var (
		query       string
		whereClause = []string{}
	)
	if request.MerchantID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("merchant_id = '%s'", request.MerchantID.String()))
	}
	if len(request.AccountIDs) > 0 {
		// / Notes:
		// / Calculate transactions aggregation with list of account ids to support wallet whitelabel merchant dashboard customer total balance
		whereClause = append(whereClause, fmt.Sprintf("account_id IN ('%s')", strings.Join(request.AccountIDs, "','")))
	}
	if request.AccountID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("account_id = '%s'", request.AccountID.String()))
	}
	if !request.IncludeFeeIndirectDeduction {
		whereClause = append(whereClause,
			`(
				type != 'FEE'
				OR (
					additional_info->>'$.deductionType' != 'MANUAL' 
					AND (
						(additional_info->>'$.deductionType' = 'DIRECT' OR additional_info->>'$.deductionType' = '')
						OR ( additional_info->>'$.deductionType' = 'AUTOMATED' AND status = 'SUCCESS' )
					)
				)
			)`,
		)
	}
	if len(whereClause) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(whereClause, " AND "))
	}

	// Handle settlement status here
	if len(request.Statuses) > 0 {
		statusesClause := []string{}
		for _, status := range request.Statuses {
			if status == constant.StatusSuccess {
				statusesClause = append(
					statusesClause,
					"(status = 'SUCCESS' and (settlement_status IS NULL OR settlement_status = 'SUCCESS'))",
				)
			} else {
				statusesClause = append(statusesClause, fmt.Sprintf("status = '%s' AND settlement_status IS NULL", status))
			}
		}

		query += fmt.Sprintf(" AND (%s)", strings.Join(statusesClause, " OR "))
	}

	if request.StartAt != nil && request.EndAt != nil {
		query += fmt.Sprintf(
			" AND (updated_at >= '%s' AND updated_at < '%s')", request.StartAt, request.EndAt,
		)
	}

	query += fmt.Sprintf(" AND (settlement_model NOT IN ('%s', '%s') OR settlement_model IS NULL)", constant.PaymentMethodChannelTypeFacilitator, constant.PaymentMethodChannelTypeDirect)

	return query
}

func (r *AccountTransactionRepository) CalculatePendingBalance(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (result float64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/CalculatePendingBalance")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		SELECT 
			COALESCE(SUM(
				CASE 
					WHEN type IN ('PAYMENT', 'MERCHANT_PAYMENT')
						 AND status = 'SUCCESS' 
						 AND settlement_status = 'PENDING' 
					THEN credit
					ELSE 0
				END
			), 0) 
			-
			COALESCE(SUM(
				CASE 
					WHEN type IN ('FEE')
						 AND status = 'SUCCESS' 
						 AND settlement_status = 'PENDING' 
					THEN debit
					ELSE 0
				END
			), 0) AS payment_pending_balance
		FROM account_transactions`

	whereClause := []string{}
	if request.MerchantID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("merchant_id = '%s'", request.MerchantID.String()))
	}
	if request.AccountID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("account_id = '%s'", request.AccountID.String()))
	}
	if request.StartAt != nil && request.EndAt != nil {
		whereClause = append(whereClause, fmt.Sprintf("(updated_at >= '%s' AND updated_at < '%s')", request.StartAt, request.EndAt))
	}
	if !request.IncludeFeeIndirectDeduction {
		whereClause = append(whereClause,
			`(
				type != 'FEE'
				OR (
					additional_info->>'$.deductionType' != 'MANUAL'
					AND (
						(additional_info->>'$.deductionType' = 'DIRECT' OR additional_info->>'$.deductionType' = '')
						OR ( additional_info->>'$.deductionType' = 'AUTOMATED' AND status = 'SUCCESS' )
					)
				)
			)`,
		)
	}

	// Force to exclude settlement model AGGREGATOR
	whereClause = append(whereClause, fmt.Sprintf("(settlement_model NOT IN ('%s', '%s') OR settlement_model IS NULL)", constant.PaymentMethodChannelTypeFacilitator, constant.PaymentMethodChannelTypeDirect))

	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}

	if err = r.db.GetContext(ctx, &result, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}

func (r *AccountTransactionRepository) GetAggregateTransactionByReference(
	ctx context.Context, request *orchestratorModel.GetSummaryTransactionByReferenceRequest) (*orchestratorModel.AggregateResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/SummaryTransactionByReference")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var aggregate orchestratorModel.AggregateResponse

	query := `
			SELECT 
				COUNT(CASE WHEN credit > 0 THEN 1 END) AS count_of_credit,
				COUNT(CASE WHEN debit > 0 THEN 1 END) AS count_of_debit,
				COALESCE(SUM(credit), 0) AS sum_of_credit, 
				COALESCE(SUM(debit), 0) AS sum_of_debit
			FROM account_transactions`

	whereClause := []string{
		fmt.Sprintf("reference_id = '%s'", request.ReferenceID),
		fmt.Sprintf("type = '%s'", request.ReferenceType),
	}
	if request.Status != "" {
		whereClause = append(whereClause, fmt.Sprintf("status = '%s'", request.Status))
	}
	if request.SettlementStatus != "" {
		whereClause = append(whereClause, fmt.Sprintf("settlement_status = '%s'", request.SettlementStatus))
	}
	if len(whereClause) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(whereClause, " AND "))
	}

	if err := r.db.GetContext(ctx, &aggregate, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when get aggregate transaction by reference", logger.Error(err))
		return nil, err
	}
	return &aggregate, nil
}

func (r *AccountTransactionRepository) GetEarliestUpdatedAt(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (time.Time, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetEarliestUpdatedAt")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	earliestUpdatedAt := time.Time{}

	query := `
			SELECT 
				COALESCE(MIN(updated_at), NOW()) as ealiest_updated_at
			FROM account_transactions`

	whereClause := []string{}
	if request.MerchantID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("merchant_id = '%s'", request.MerchantID.String()))
	}
	if len(request.AccountIDs) > 0 {
		// / Notes:
		// / Calculate transactions aggregation with list of account ids to support wallet whitelabel merchant dashboard customer total balance
		whereClause = append(whereClause, fmt.Sprintf("account_id IN ('%s')", strings.Join(request.AccountIDs, "','")))
	}
	if request.AccountID != uuid.Nil {
		whereClause = append(whereClause, fmt.Sprintf("account_id = '%s'", request.AccountID.String()))
	}
	if request.PendingSettlementBalance {
		whereClause = append(whereClause, "status = 'SUCCESS' AND IFNULL(settlement_status, 'SUCCESS') = 'PENDING'")
	}
	if !request.IncludeFeeIndirectDeduction {
		whereClause = append(whereClause,
			`(
				type != 'FEE'
				OR (
					additional_info->>'$.deductionType' != 'MANUAL' 
					AND (
						(additional_info->>'$.deductionType' = 'DIRECT' OR additional_info->>'$.deductionType' = '')
						OR ( additional_info->>'$.deductionType' = 'AUTOMATED' AND status = 'SUCCESS' )
					)
				)
			)`,
		)
	}
	if len(whereClause) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(whereClause, " AND "))
	}

	// Handle settlement status here
	if len(request.Statuses) > 0 {
		statusesClause := []string{}
		for _, status := range request.Statuses {
			if status == constant.StatusSuccess {
				statusesClause = append(
					statusesClause,
					"(status = 'SUCCESS' and (settlement_status IS NULL OR settlement_status = 'SUCCESS'))",
				)
			} else {
				statusesClause = append(statusesClause, fmt.Sprintf("status = '%s' AND settlement_status IS NULL", status))
			}
		}

		query += fmt.Sprintf(" AND (%s)", strings.Join(statusesClause, " OR "))
	}

	if request.StartAt != nil && request.EndAt != nil {
		query += fmt.Sprintf(
			" AND (updated_at >= '%s' AND updated_at < '%s')", request.StartAt, request.EndAt,
		)
	}

	query += fmt.Sprintf(" AND (settlement_model NOT IN ('%s', '%s') OR settlement_model IS NULL)", constant.PaymentMethodChannelTypeFacilitator, constant.PaymentMethodChannelTypeDirect)
	if err := r.db.GetContext(ctx, &earliestUpdatedAt, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, nil
		}

		r.logger.Error(ctx, "error when get earliest updated at", logger.Error(err))
		return time.Time{}, err
	}
	return earliestUpdatedAt, nil
}

func (r *AccountTransactionRepository) GetBulkAggregateTransactions(ctx context.Context, request *orchestratorModel.BulkGetAggregateRequest) ([]*orchestratorModel.BulkAggregateResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetBulkAggregateTransactions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var (
		aggregates           []*orchestratorModel.BulkAggregateResponse
		whereClause          = []string{}
		accountParentClauses = make([]string, len(request.AccountClauses))
	)

	query := `
			SELECT 
				merchant_id AS merchant_id, 
				account_id AS account_id,
				COUNT(CASE WHEN credit > 0 THEN 1 END) AS count_of_credit,
				COUNT(CASE WHEN debit > 0 THEN 1 END) AS count_of_debit,
				COALESCE(SUM(credit), 0) AS sum_of_credit, 
				COALESCE(SUM(debit), 0) AS sum_of_debit
			FROM account_transactions`

	for i, clause := range request.AccountClauses {
		var accountClause []string
		if clause.MerchantID != "" {
			accountClause = append(accountClause, fmt.Sprintf("merchant_id = '%s'", clause.MerchantID))
		}
		if clause.AccountID != "" {
			accountClause = append(accountClause, fmt.Sprintf("account_id = '%s'", clause.AccountID))
		}
		if clause.StartAt != nil && clause.EndAt != nil {
			accountClause = append(accountClause, fmt.Sprintf("(updated_at >= '%s' AND updated_at < '%s')", clause.StartAt, clause.EndAt))
		}

		if len(accountClause) > 0 {
			accountParentClauses[i] = fmt.Sprintf("(%s)", strings.Join(accountClause, " AND "))
		}
	}
	if len(accountParentClauses) > 0 {
		whereClause = append(whereClause, fmt.Sprintf(" (%s) ", strings.Join(accountParentClauses, " OR ")))
	}

	if !request.IncludeFeeIndirectDeduction {
		whereClause = append(whereClause,
			`(
				type != 'FEE'
				OR (
				
					additional_info->>'$.deductionType' != 'MANUAL' 
					AND (
						(additional_info->>'$.deductionType' = 'DIRECT' OR additional_info->>'$.deductionType' = '')
						OR ( additional_info->>'$.deductionType' = 'AUTOMATED' AND status = 'SUCCESS' )
					)
				)
			)`,
		)
	}

	// Handle settlement status here
	if len(request.Statuses) > 0 {
		statusesClause := []string{}
		for _, status := range request.Statuses {
			if status == constant.StatusSuccess {
				statusesClause = append(
					statusesClause,
					"(status = 'SUCCESS' and (settlement_status IS NULL OR settlement_status = 'SUCCESS'))",
				)
			} else {
				statusesClause = append(statusesClause, fmt.Sprintf("status = '%s' AND settlement_status IS NULL", status))
			}
		}

		whereClause = append(whereClause, fmt.Sprintf(" (%s)", strings.Join(statusesClause, " OR ")))
	}

	whereClause = append(whereClause, fmt.Sprintf(" (settlement_model NOT IN ('%s', '%s') OR settlement_model IS NULL)", constant.PaymentMethodChannelTypeFacilitator, constant.PaymentMethodChannelTypeDirect))
	if len(whereClause) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(whereClause, " AND "))
	}

	query = query + " GROUP BY merchant_id,account_id"

	if err := r.db.SelectContext(ctx, &aggregates, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error(ctx, "error when get bulk aggregate transaction", logger.Error(err))
		return nil, err
	}
	return aggregates, nil
}
