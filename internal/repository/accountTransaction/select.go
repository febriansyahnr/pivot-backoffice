package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) GetDetailById(ctx context.Context, merchantId, id string) (*orchestrator_model.TransactionHistoryDetailResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetDetailById")
	defer segment.End()

	resp := &orchestrator_model.TransactionHistoryDetailResp{}
	ctx1 := context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
		uuid AS id, merchant_id, created_at, updated_at, channel, remarks, 
		CASE   
			WHEN type = 'FEE' THEN IFNULL(CONCAT(additional_info->>'$.type', '_FEE'), 'FEE')
			WHEN type = 'WITHDRAWAL' THEN CONCAT(TRIM(reference), '_', TRIM(type))
			ELSE type
		END AS type,
		CASE
			WHEN debit > 0 THEN debit
			WHEN COALESCE(credit, 0) = 0
				AND NULLIF(additional_info->>'$.subPaymentSummary.totalCreditAmount','') IS NOT NULL
			THEN CAST(
				additional_info->>'$.subPaymentSummary.totalCreditAmount'
				AS DECIMAL(18,2)
			)
			ELSE COALESCE(credit, 0)
		END AS amount, 
		status, 
		IF(status = 'FAILED' AND reason_type = 'INVALID_ACCOUNT', reason_description, '') AS reason_description,
		IF(type IN ('DISBURSEMENT', 'BULK_DISBURSEMENT', 'WITHDRAWAL', 'FEE', 'PAYMENT', 'TRANSFER'), reference_id, '') AS reference_id,
		IFNULL(additional_info->>'$.linked_transaction_id', '') AS linked_transaction_id
	FROM account_transactions
	WHERE uuid = ? AND merchant_id = ?;`
	if err := r.db.GetContext(ctx1, resp, rawQuery, id, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if resp.ReferenceId == "" {
		return resp, nil
	}

	var errGetDetail error
	switch {
	case resp.Type == constant.TypeDisbursement:
		resp.Details, errGetDetail = r.getTransactionDisbursementDetail(ctx, resp.ReferenceId)

	case strings.HasSuffix(resp.Type, constant.TypeWithdrawal):
		resp.Details, errGetDetail = r.getTransactionWithdrawalDetail(ctx, resp.ReferenceId)

	case strings.Contains(resp.Type, constant.TypeFee) && resp.LinkedTransactionId != "":
		resp.Details = &orchestrator_model.TransactionFeeResp{
			LinkedID: resp.LinkedTransactionId,
		}
	case resp.Type == constant.TypePayment:
		resp.Details, errGetDetail = r.getTransactionPaymentDetail(ctx, resp.ReferenceId)

	case resp.Type == constant.TypeTransfer:
		resp.Details, errGetDetail = r.getTransactionTransferDetail(ctx, resp.ReferenceId, resp.MerchantID)
	}

	if errGetDetail != nil {
		return nil, errGetDetail
	}

	return resp, nil
}

func (r *AccountTransactionRepository) getTransactionDisbursementDetail(ctx context.Context, disbursementID string) (*orchestrator_model.TransactionDisbursementResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/getTransactionDisbursementDetail")
	defer segment.End()

	disbursement := &orchestrator_model.TransactionDisbursementResp{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements,users")

	rawQuery := `SELECT
			IFNULL(uc.name, m.name) AS created_by, d.reference_id, d.bank_reference_no, d.beneficiary_bank_name, d.beneficiary_account_no,
			d.bulk_id,d.beneficiary_account_name, d.fee, d.total_amount, d.status, d.reason_type, d.reason_description, d.approved_at, IFNULL(ua.name, 'AUTO') AS approved_by, d.sender_name, d.created_from, d.currency
		FROM disbursements d 
		LEFT JOIN users uc ON uc.uuid = d.created_by 
		LEFT JOIN users ua ON ua.uuid = d.approved_by
		LEFT JOIN merchants m ON m.uuid = d.merchant_id
		WHERE d.uuid = ?;`
	if err := r.db.GetContext(ctx, disbursement, rawQuery, disbursementID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return disbursement, nil
}

func (r *AccountTransactionRepository) getTransactionWithdrawalDetail(ctx context.Context, withdrawalID string) (*orchestrator_model.TransactionWithdrawalResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/getTransactionWithdrawalDetail")
	defer segment.End()

	withdrawal := &orchestrator_model.TransactionWithdrawalResp{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "withdrawals,users")

	rawQuery := `SELECT
			IFNULL(uc.name, 'System') AS created_by, 
			IFNULL(metadata->>'$.bankTransfer.bankReferenceNo', '') AS bank_reference_no,
			w.beneficiary_bank_name, w.beneficiary_account_no, w.beneficiary_account_name 
		FROM withdrawals w 
		LEFT JOIN users uc ON uc.uuid = w.created_by 
		WHERE w.id = ?;`
	if err := r.db.GetContext(ctx, withdrawal, rawQuery, withdrawalID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return withdrawal, nil
}

func (r *AccountTransactionRepository) GetPlatformTransactionActivities(ctx context.Context, ids []string, startDate, endDate time.Time) (result []orchestrator_model.TransactionActivity, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetPlatformTransactionActivities")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
		merchant_id, DATE_FORMAT(CONVERT_TZ(updated_at, '+00:00', '+07:00'), '%Y%m') AS period, COUNT(uuid) AS total
	FROM account_transactions 
	WHERE 
		merchant_id IN (?) AND (updated_at BETWEEN ? AND ?) AND type != 'FEE' AND status = 'SUCCESS'
	GROUP BY merchant_id, period;`
	query, args, err := r.db.In(rawQuery, ids, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed generate in statement: %v", err)
	}
	query = r.db.Rebind(query)

	if err = r.db.SelectContext(ctx, &result, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
	}
	return
}

func (r *AccountTransactionRepository) GetAccumulateTransactionFees(ctx context.Context, merchantId, reference, method string, startDate, endDate time.Time) (*orchestrator_model.AccumulateTransactionFees, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetAccumulateTransactionFees")
	defer segment.End()

	result := &orchestrator_model.AccumulateTransactionFees{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
			COUNT(uuid) AS total_rows, IFNULL(SUM(debit), 0) AS total_fees,
			IFNULL(SUM(additional_info->>'$.taxAmount'), 0) AS total_taxes, IFNULL(JSON_ARRAYAGG(uuid), '[]') AS transaction_ids,
			IFNULL((SELECT acc.name FROM accounts acc WHERE acc.uuid = GROUP_CONCAT(DISTINCT at.account_id)), '') AS account_name
		FROM account_transactions at
		WHERE merchant_id = ? AND (updated_at BETWEEN ? AND ?)
			AND type = 'FEE' AND additional_info->>'$.type' = ? AND IFNULL(additional_info->>'$.method', '') = ?
			AND additional_info->>'$.deductionType' = 'AUTOMATED' 
			AND additional_info->>'$.canceledAt' IS NULL AND status = 'PENDING'
			AND (settlement_status IS NULL OR settlement_status = 'SUCCESS');`
	if err := r.db.GetContext(ctx, result, rawQuery, merchantId, startDate, endDate, reference, method); err != nil {
		return nil, err
	}

	result.RawTransactionIds.Unmarshal(&result.TransactionIds)
	return result, nil
}

func (r *AccountTransactionRepository) CalculatingMerchantTPVToDetermineFeeTierLevel(ctx context.Context, merchantId string, startDate, endDate time.Time) (result map[string]orchestrator_model.CalculatingMerchantTPVSummary, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/CalculatingMerchantTPVToDetermineFeeTierLevel")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
			type, channel, additional_info->>'$.type' AS additional, 
			COUNT(uuid) AS frequency, SUM(debit+credit) AS volume
		FROM 
			account_transactions
		WHERE merchant_id = ?
			AND (updated_at BETWEEN ? AND ?)
			AND (type != 'FEE' OR (type = 'FEE' AND additional_info->>'$.type' = 'ACCOUNT_INQUIRY'))
			AND status = 'SUCCESS' 
			AND IFNULL(settlement_status, 'SUCCESS') = 'SUCCESS'
			AND channel NOT IN ('MANUAL_TRANSFER', 'BALANCE_ADJUSTMENT', 'MANUAL_ACTION')
		GROUP BY 
			type, channel, additional 
		ORDER BY 
			type, channel, additional;`
	tpvSummaries := []orchestrator_model.CalculatingMerchantTPVSummary{}

	if err = r.db.SelectContext(ctx, &tpvSummaries, rawQuery, merchantId, startDate, endDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result = map[string]orchestrator_model.CalculatingMerchantTPVSummary{}

	for _, summary := range tpvSummaries {

		key := summary.Type

		if summary.Additional != nil {
			key = *summary.Additional

		} else if summary.Channel != "" {
			key += "_" + summary.Channel
		}

		result[key] = summary
	}
	return
}

// CalculatingMerchantTPVForLadderTiering counts all successful transactions
func (r *AccountTransactionRepository) CalculatingMerchantTPVForLadderTiering(ctx context.Context, merchantId string, startDate, endDate time.Time) (result map[string]orchestrator_model.CalculatingMerchantTPVSummary, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/CalculatingMerchantTPVForLadderTiering")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
			type, channel, additional_info->>'$.type' AS additional,
			COUNT(uuid) AS frequency, SUM(debit+credit) AS volume
		FROM
			account_transactions
		WHERE merchant_id = ?
			AND (updated_at BETWEEN ? AND ?)
			AND (type != 'FEE' OR (type = 'FEE' AND additional_info->>'$.type' = 'ACCOUNT_INQUIRY'))
			AND status = 'SUCCESS'
			AND channel NOT IN ('MANUAL_TRANSFER', 'BALANCE_ADJUSTMENT', 'MANUAL_ACTION')
		GROUP BY
			type, channel, additional
		ORDER BY
			type, channel, additional;`
	tpvSummaries := []orchestrator_model.CalculatingMerchantTPVSummary{}

	if err = r.db.SelectContext(ctx, &tpvSummaries, rawQuery, merchantId, startDate, endDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result = map[string]orchestrator_model.CalculatingMerchantTPVSummary{}

	for _, summary := range tpvSummaries {
		key := summary.Type

		if summary.Additional != nil {
			key = *summary.Additional
		} else if summary.Channel != "" {
			key += "_" + summary.Channel
		}

		result[key] = summary
	}
	return
}

func (r *AccountTransactionRepository) CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel(ctx context.Context, merchantIds []string, startDate, endDate time.Time) (map[string]orchestrator_model.CalculatingMerchantTPVSummary, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	rawQuery := `
		SELECT
			'PLATFORM_ACTIVITY' AS type, COUNT(uuid) AS frequency, IFNULL(SUM(debit+credit), 0) AS volume
		FROM 
			account_transactions
		WHERE merchant_id IN (?)
			AND (updated_at BETWEEN ? AND ?)
			AND (type != 'FEE' OR (type = 'FEE' AND additional_info->>'$.type' = 'ACCOUNT_INQUIRY'))
			AND status = 'SUCCESS' 
			AND IFNULL(settlement_status, 'SUCCESS') = 'SUCCESS'
			AND channel NOT IN ('MANUAL_TRANSFER', 'BALANCE_ADJUSTMENT', 'MANUAL_ACTION');`
	query, args, err := r.db.In(rawQuery, merchantIds, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("in statement: %w", err)
	}

	tpvSummary := orchestrator_model.CalculatingMerchantTPVSummary{}
	if err = r.db.GetContext(ctx, &tpvSummary, query, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return map[string]orchestrator_model.CalculatingMerchantTPVSummary{
		constant.ReferencePlatformActivity: tpvSummary,
	}, nil
}

func (r *AccountTransactionRepository) getTransactionPaymentDetail(ctx context.Context, paymentId string) (*orchestrator_model.TransactionPaymentResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/getTransactionPaymentDetail")
	defer segment.End()

	payment := &orchestrator_model.TransactionPaymentResp{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments,payment_methods")

	rawQuery := `SELECT
			pm.type, pm.name, pm.category, pm.description 
		FROM payments p
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid 
		WHERE p.uuid = ?;`
	if err := r.db.GetContext(ctx, payment, rawQuery, paymentId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return payment, nil
}

func (r *AccountTransactionRepository) getTransactionTransferDetail(ctx context.Context, transferId, merchantId string) (*orchestrator_model.TransactionTransferResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/getTransactionTransferDetail")
	defer segment.End()

	transfer := &orchestrator_model.TransactionTransferResp{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments,payment_methods")

	rawQuery := `SELECT
			merchant_id, recipient_id, currency, amount 
		FROM transfers  
		WHERE uuid = ?;`
	if err := r.db.GetContext(ctx, transfer, rawQuery, transferId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	transfer.Type = constant.TransferTypeIN
	if transfer.MerchantID == merchantId {
		transfer.Type = constant.TransferTypeOUT
	}

	return transfer, nil
}

func (r *AccountTransactionRepository) FindLastMerchantTransactionDate(ctx context.Context, merchantID string) (*time.Time, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/FindLastMerchantTransactionDate")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	data := time.Time{}
	query := `
		SELECT updated_at FROM account_transactions WHERE merchant_id = ? ORDER BY updated_at DESC LIMIT 1;
	`

	if err := r.db.GetContext(ctx, &data, query, merchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find last merchant transaction", logger.Error(err))
		return nil, err
	}

	return &data, nil
}

func (r *AccountTransactionRepository) GetLastTransactionByAccountName(ctx context.Context, merchantId, accountName string) (result *time.Time, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetLastTransactionByAccountType")
	defer segment.End()

	result, ctx = &time.Time{}, context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	rawQuery := `SELECT MAX(updated_at) FROM account_transactions WHERE merchant_id = ? AND reference = ?;`

	err = r.db.GetContext(ctx, result, rawQuery, merchantId, accountName)
	return
}

// For now, we only need the UUID, but it is possible that there is other data.
func (r *AccountTransactionRepository) GetTransactionByReferenceIdAndProcessorId(ctx context.Context, referenceId, processorId string) (*orchestrator_model.AccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetTransactionByReferenceIdAndProcessorId")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT uuid, settlement_model FROM account_transactions WHERE reference_id = ? AND processor_reference_id = ?;`

	result := orchestrator_model.AccountTransaction{}
	if err := r.db.GetContext(ctx, &result, rawQuery, referenceId, processorId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (r *AccountTransactionRepository) GetReferenceIdByTransactionIdAndType(ctx context.Context, transactionId, transactionType string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetReferenceIdByTransactionIdAndType")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT reference_id FROM account_transactions WHERE uuid = ? AND type = ?;`

	result := ""
	if err := r.db.GetContext(ctx, &result, rawQuery, transactionId, transactionType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return result, nil
}
