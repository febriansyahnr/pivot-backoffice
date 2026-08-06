package reportingRepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *repository) UpsertBalanceHistory(ctx context.Context, data model.BalanceHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/UpsertBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	rawQuery := `INSERT INTO report_balance_histories (
		transaction_id, merchant_id, type, reference_id, balance_type, channel, transaction_type, currency, amount, fee, 
		remarks, status, reason_type, reason_description, settlement_model, settlement_status, settlement_at, additional_info, created_at, status_updated_at,
		source_id, source_account_id, source_created_at, source_created_by, _ingested_at
	)
	VALUES(
		:transaction_id, :merchant_id, :type, :reference_id, :balance_type, :channel, :transaction_type, :currency, :amount, :fee, 
		:remarks, :status, :reason_type, :reason_description, :settlement_model, :settlement_status, :settlement_at, :additional_info, :created_at, :status_updated_at,
		:source_id, :source_account_id, :source_created_at, :source_created_by, :_ingested_at
	) ON DUPLICATE KEY UPDATE type = VALUES(type), channel = VALUES(channel), transaction_type = VALUES(transaction_type),
	 	amount = VALUES(amount), fee = VALUES(fee), status = VALUES(status), reason_type = VALUES(reason_type),
	 	reason_description = VALUES(reason_description), status_updated_at = VALUES(status_updated_at), settlement_status = VALUES(settlement_status), settlement_at = VALUES(settlement_at),
		additional_info = VALUES(additional_info), _ingested_at = VALUES(_ingested_at);`

	if _, err := r.db.NamedExecContext(ctx, rawQuery, data); err != nil {
		r.logger.Error(ctx, "Failed to upsert balance history data", logger.Error(err))
		return err
	}
	return nil
}

func (r *repository) HardDeleteBalanceHistory(ctx context.Context, transactionID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/HardDeleteBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	rawQuery := `DELETE FROM report_balance_histories WHERE transaction_id = ?;`
	if _, err := r.db.ExecContext(ctx, rawQuery, transactionID); err != nil {
		r.logger.Error(ctx, "Failed to hard delete balance history data", logger.Error(err))
		return err
	}
	return nil
}

func (r *repository) SoftDeleteBalanceHistory(ctx context.Context, transactionID string, ingestedAt time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/SoftDeleteBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	rawQuery := `UPDATE report_balance_histories SET _is_deleted = 1, _deleted_at = ?, _ingested_at = ? WHERE transaction_id = ?;`
	if _, err := r.db.ExecContext(ctx, rawQuery, time.Now().UTC(), ingestedAt, transactionID); err != nil {
		r.logger.Error(ctx, "Failed to soft delete balance history data", logger.Error(err))
		return err
	}
	return nil
}

func (r *repository) UpdateSettlementBalanceHistory(ctx context.Context, data model.BalanceHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/UpdateSettlementBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	rawQuery := `UPDATE
		report_balance_histories
	SET
		status_updated_at = :status_updated_at, settlement_status = :settlement_status, settlement_at = :settlement_at, _ingested_at = :_ingested_at
	WHERE
		transaction_id = :transaction_id;`
	if effected, err := r.db.NamedExecContext(ctx, rawQuery, data); err != nil {
		r.logger.Error(ctx, "Failed to update settlement balance history data", logger.Error(err))
		return err
	} else if !effected {
		return constant.ErrNoRowsAffected
	}
	return nil
}

func (r *repository) PrepareAdvancedBalanceHistoryData(ctx context.Context, data *model.BalanceHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/PrepareAdvancedBalanceHistoryData")
	defer segment.End()

	tableName, rawQuery, args := getAdvancedBalanceHistoryQuery(data)

	if tableName == "" {
		return nil
	}

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	if err := r.db.GetContext(ctx, data, rawQuery, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.logger.Warn(ctx,
			"Failed to run the balance history data preparation query",
			logger.String("rawQuery", strings.ReplaceAll(strings.ReplaceAll(rawQuery, "\n", ""), "\t", "")), logger.Any("args", args),
		)
		return err
	}
	return nil
}

func getAdvancedBalanceHistoryQuery(data *model.BalanceHistory) (tableName, rawQuery string, args []any) {

	switch data.Type {
	case constant.TypeDisbursement:
		return "disbursements", `SELECT 
				reference_id, IF(bulk_id IS NOT NULL, 'BULK_DISBURSEMENT', 'DISBURSEMENT') AS transaction_type, IFNULL(d.metadata->>'$.feeDetail.finalAmount', d.fee) AS fee,
				d.created_at AS source_created_at, IFNULL(IFNULL(m.name, u.name), mt.name) AS source_created_by, 
				JSON_OBJECT(
					'bankReferenceNo', d.bank_reference_no, 'beneficiaryBankName', d.beneficiary_bank_name, 'beneficiaryAccountNo', d.beneficiary_account_no, 'beneficiaryName', d.beneficiary_account_name
				) AS additional_info 
			FROM disbursements d
			LEFT JOIN merchants m ON m.uuid = d.created_by
			LEFT JOIN users u ON u.uuid = d.created_by 
			LEFT JOIN merchants mt ON mt.uuid = d.merchant_id
			WHERE 
				d.uuid = ?;`, []any{data.SourceID}

	case constant.TypePayment:
		return "payments", `SELECT
				reference_id, IF(? > 0, ?, IFNULL(p.metadata->>'$.feeDetail.finalAmount', 0)) AS fee, p.created_at AS source_created_at, IFNULL(m.name, u.name) AS source_created_by
			FROM 
				payments p
			LEFT JOIN merchants m ON m.uuid = p.created_by
			LEFT JOIN users u ON u.uuid = p.created_by
			WHERE 
				p.uuid = ?;`, []any{data.Fee, data.Fee, data.SourceID}

	case constant.TypeWithdrawal:
		return "withdrawals", `SELECT
				w.created_at AS source_created_at, IFNULL(IFNULL(m.name, u.name),'System') AS source_created_by,
				IF(w.metadata->>'$.withdrawType' = 'BANK_TRANSFER', JSON_OBJECT(
					'bankReferenceNo', w.metadata->>'$.bankTransfer.bankReferenceNo', 'beneficiaryBankName', w.beneficiary_bank_name, 'beneficiaryAccountNo', w.beneficiary_account_no, 'beneficiaryName', w.beneficiary_account_name
				), NULL) AS additional_info
			FROM
				withdrawals w
			LEFT JOIN merchants m ON m.uuid = w.created_by 
			LEFT JOIN users u ON u.uuid = w.created_by 
			WHERE
				w.id = ?;`, []any{data.SourceID}

	case constant.TypeTransfer:
		return "transfers", `SELECT
				t.reference_id, t.created_at AS source_created_at, IF(t.remarks LIKE '%Fee Transfer%', 'System', m.name) AS source_created_by, IFNULL(at.debit, 0) AS fee
			FROM transfers t
			JOIN merchants m ON m.uuid = t.merchant_id
			LEFT JOIN account_transactions at ON at.reference_id = t.uuid 
				AND at.merchant_id = ?
				AND (at.created_at BETWEEN DATE_SUB(?, INTERVAL 3 MINUTE) AND DATE_ADD(?, INTERVAL 3 MINUTE)) AND at.type = 'FEE'
			WHERE
				t.uuid = ?;`, []any{data.MerchantID, data.CreatedAt, data.CreatedAt, data.SourceID}

	case constant.TypeRefund:
		return "refunds", `SELECT
				r.client_reference_id AS reference_id, r.created_at AS source_created_at, m.name AS source_created_by, IFNULL(at.debit, 0) AS fee,
				JSON_OBJECT('paymentId', r.payment_id, 'paymentChargeId', r.payment_charge_id, 'paymentReferenceId', p.reference_id) AS additional_info
			FROM refunds r
			JOIN merchants m ON m.uuid = r.merchant_id
			JOIN payments p ON p.uuid = r.payment_id
			LEFT JOIN account_transactions at ON at.reference_id = r.uuid
				AND at.merchant_id = ?
				AND (at.created_at BETWEEN DATE_SUB(?, INTERVAL 3 MINUTE) AND DATE_ADD(?, INTERVAL 3 MINUTE)) AND at.type = 'FEE'
			WHERE
				r.uuid = ?;`, []any{data.MerchantID, data.CreatedAt, data.CreatedAt, data.SourceID}

	case constant.TypeVirtualTerminal:
		tmp := *data
		tmp.Type = constant.TypePayment
		return getAdvancedBalanceHistoryQuery(&tmp)

	case constant.TypeGeneralTopUp:
		return "account_transactions", `SELECT
				debit AS fee
			FROM account_transactions
			WHERE
				reference_id = ? AND merchant_id = ? AND type = 'FEE' 
				AND additional_info->>'$.referenceType' = ? AND debit > 0 
				AND (created_at BETWEEN DATE_SUB(?, INTERVAL 3 MINUTE) AND DATE_ADD(?, INTERVAL 3 MINUTE))`,
			[]any{data.SourceID, data.MerchantID, constant.TypeGeneralTopUp, data.CreatedAt, data.CreatedAt}

	case constant.TypeMerchantPayment, constant.TypeCashback:
		return "merchants", "SELECT name AS source_created_by FROM merchants WHERE uuid = ?;", []any{data.MerchantID}

	case constant.TypeManualAdjust:
		return "manual_adjustment_histories", `SELECT
				reference_id, created_at AS source_created_at, 'Pivot Ops' AS source_created_by, IF(type = 'DISBURSEMENT' AND action = 'NORMAL', JSON_OBJECT('bankReferenceNo', bank_reference_id), NULL) AS additional_info
			FROM manual_adjustment_histories
			WHERE uuid = ?;`, []any{data.SourceID}
	}

	if data.Type == constant.TypeReversal && data.TransactionType == constant.TypeDisbursement {
		return "disbursements", `SELECT
				reference_id, IF(bulk_id IS NOT NULL, 'BULK_DISBURSEMENT', 'DISBURSEMENT') AS transaction_type, 'Pivot Ops' AS source_created_by,
				JSON_OBJECT(
					'bankReferenceNo', bank_reference_no, 'beneficiaryBankName', beneficiary_bank_name, 'beneficiaryAccountNo', beneficiary_account_no, 'beneficiaryName', beneficiary_account_name
				) AS additional_info
			FROM disbursements
			WHERE uuid = ?;`, []any{data.SourceID}
	}

	if data.TransactionType == constant.TypeDisbursementFee {
		return "disbursements", `SELECT
				reference_id, created_at AS source_created_at,
				JSON_OBJECT(
					'bankReferenceNo', bank_reference_no, 'beneficiaryBankName', beneficiary_bank_name, 'beneficiaryAccountNo', beneficiary_account_no, 'beneficiaryName', beneficiary_account_name
				) AS additional_info
			FROM disbursements
			WHERE uuid = ?;`, []any{data.SourceID}
	}

	return "", "", nil
}

func (r *repository) ListBalanceHistory(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/ListBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	// Initialize pagination utility
	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      r.appConfig.InitialPageWindow,
	}
	paginationUtil := util.NewPaginationUtility(r.db, r.logger, paginationConfig)

	// Query components
	queryBuilder := util.QueryBuilder{
		SelectQuery: balanceHistorySelectQuery,
		CountQuery:  "SELECT COUNT(transaction_id) FROM report_balance_histories",
	}

	// Filter conditions
	filterResult := util.FilterResult{}
	filterResult.Conditions, filterResult.Args = buildConditionForBalanceHistory(filter)

	// Sort configuration
	sortConfig := util.SortConfig{DefaultSort: filter.FilteredSortQuery}

	// Data destination
	data := make([]orchestratorModel.AccountTransactionWithUseCase, 0)

	// Data transformer
	featureName, _ := ctx.Value(constant.CtxFeatureName).(string)
	dataTransformer := func(result any) any {
		transactions := *result.(*[]orchestratorModel.AccountTransactionWithUseCase)

		dest := make([]any, len(transactions))
		for i, transaction := range transactions {
			if featureName == constant.FeatureBalanceHistoryOpenApi {
				dest[i] = orchestratorModel.ToTransactionHistoryOpenApiResponse(&transaction)
			} else {
				dest[i] = orchestratorModel.ToTransactionHistoryResponse(&transaction)
			}
		}
		return dest
	}
	return paginationUtil.GetPaginatedList(
		ctx, queryBuilder, filterResult, sortConfig, page, perPage, &data, dataTransformer,
	)
}

func (r *repository) ExportBalanceHistory(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest) ([]orchestratorModel.TransactionHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/reporting/ExportBalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, r.TableName())

	conditions, args := buildConditionForBalanceHistory(filter)
	rawQuery := balanceHistorySelectQuery + " WHERE " + strings.Join(conditions, " AND ") + " ORDER BY " + filter.FilteredSortQuery

	transactions := make([]orchestratorModel.AccountTransactionWithUseCase, 0)
	if err := r.db.SelectContext(ctx, &transactions, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := make([]orchestratorModel.TransactionHistory, len(transactions))
	for i, transaction := range transactions {
		result[i] = transaction.ToTransactionHistory()
	}
	return result, nil
}
