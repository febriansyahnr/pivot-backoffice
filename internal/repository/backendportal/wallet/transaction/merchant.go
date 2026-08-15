package walletTransaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/wallet/transaction"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"golang.org/x/sync/errgroup"
)

const (
	merchantTransactionHistoryColums = `at.uuid, at.type, IF(at.channel != '', at.channel, IFNULL(at.additional_info->>'$.referenceType', '')) AS channel, at.created_at, 
		at.updated_at, IF(at.credit > 0, at.credit, -at.debit) AS amount, at.status, 
		IF(at.status != 'SUCCESS', at.status, IFNULL(at.settlement_status, 'SUCCESS')) AS settlement_status, 
		IF(at.type IN ('MERCHANT_TOP_UP','WITHDRAWAL'), '-', IFNULL(at.merchant_reference_id, '')) AS merchant_reference_id`
	merchantTransactionHistoryCountRows = "IFNULL(COUNT(uuid), 0)"
)

var merchantTransactionHistoryOrderColumn = map[string]string{
	"date": "updated_at ASC", "-date": "updated_at DESC",
}

func (repository) buildConditionAndArgsTransactionHistoryList(req walletTransactionModel.MerchantTransactionHistoryListReq) ([]string, []any) {
	conditions, args := []string{}, []any{req.MerchantId, req.StartDate, req.EndDate}

	if req.Type != "" {
		if strings.HasPrefix(req.Type, constant.TypeFee) {
			// Example: FEE_BILL_PAYMENT or FEE_WITHDRAWAL
			args = append(args, constant.TypeFee, strings.Replace(req.Type, "FEE_", "", 1))
			conditions = append(conditions, "(at.type = ? AND IF(at.channel != '', at.channel, IFNULL(at.additional_info->>'$.referenceType', '')) = ?)")
		} else {
			// Example: MERCHANT_PAYMENT
			args = append(args, req.Type)
			conditions = append(conditions, "at.type = ?")
		}
	}
	if req.Status != "" {
		args = append(args, req.Status)
		conditions = append(conditions, "IF(at.status != 'SUCCESS', at.status, IFNULL(at.settlement_status, 'SUCCESS')) = ?")
	}
	if req.Id != "" {
		args = append(args, req.Id)
		conditions = append(conditions, "at.uuid = ?")
	}
	if req.ReferenceId != "" {
		args = append(args, "%"+req.ReferenceId+"%")
		conditions = append(conditions, "at.merchant_reference_id LIKE ?")
	}
	return conditions, args
}

func (r *repository) GetMerchantTransactionHistoryList(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) ([]walletTransactionModel.MerchantTransactionHistoryListResp, int64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/wallet/transaction/GetMerchantTransactionHistoryList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT %s
	FROM account_transactions at
	WHERE
		merchant_id = ?
		AND (updated_at BETWEEN ? AND ?)
		AND account_id = (SELECT uuid FROM accounts WHERE reference_id = at.merchant_id AND name = 'WALLET')
		AND NOT (status = 'PENDING' AND settlement_status = 'PENDING')`

	conditions, args := r.buildConditionAndArgsTransactionHistoryList(req)

	var (
		totalRows = int64(0)
		errGroups = new(errgroup.Group)
		result    = []walletTransactionModel.MerchantTransactionHistoryListResp{}
	)
	if len(conditions) > 0 {
		rawQuery = rawQuery + " AND " + strings.Join(conditions, " AND ")
	}

	errGroups.Go(func() error {
		return r.db.GetContext(
			ctx, &totalRows, fmt.Sprintf(rawQuery, merchantTransactionHistoryCountRows), args...,
		)
	})

	errGroups.Go(func() error {
		query := fmt.Sprintf(rawQuery, merchantTransactionHistoryColums) +
			" ORDER BY " + merchantTransactionHistoryOrderColumn[req.Sort] + " LIMIT ? OFFSET ?"
		return r.db.SelectContext(ctx, &result, query, append(args, req.PerPage, ((req.Page-1)*req.PerPage))...)
	})

	if err := errGroups.Wait(); err != nil {
		return nil, 0, err
	}
	return result, totalRows, nil
}

func (r *repository) GetMerchantTransactionHistoryListForExport(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) (result []walletTransactionModel.MerchantTransactionHistoryListResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/wallet/transaction/GetMerchantTransactionHistoryListForExport")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
		` + merchantTransactionHistoryColums + `,
		CASE
			WHEN at.type = 'MERCHANT_PAYMENT' THEN cus.first_name
			WHEN at.type = 'WITHDRAWAL' THEN w_user.name
			ELSE 'System'
		END AS created_by, IFNULL(w.beneficiary_account_no, '') AS beneficiary_account_no,
		IFNULL(w.beneficiary_account_name, '') AS beneficiary_account_name
	FROM account_transactions at
	LEFT JOIN account_transactions at_cus ON at_cus.reference_id = at.reference_id AND at_cus.type = 'MERCHANT_PAYMENT' AND at_cus.account_id != at.account_id
	LEFT JOIN accounts acc_cus ON acc_cus.uuid = at_cus.account_id
	LEFT JOIN customers cus ON cus.uuid = acc_cus.reference_id
	LEFT JOIN withdrawals w ON w.id = at.reference_id AND at.type = 'WITHDRAWAL'
	LEFT JOIN users w_user ON w_user.uuid = w.created_by
	WHERE
		at.merchant_id = ?
		AND (at.updated_at BETWEEN ? AND ?)
		AND at.account_id = (SELECT uuid FROM accounts WHERE reference_id = at.merchant_id AND name = 'WALLET')
		AND NOT (at.status = 'PENDING' AND at.settlement_status = 'PENDING')`

	conditions, args := r.buildConditionAndArgsTransactionHistoryList(req)
	if len(conditions) > 0 {
		rawQuery += " AND " + strings.Join(conditions, " AND ")
	}

	rawQuery += " ORDER BY " + merchantTransactionHistoryOrderColumn[req.Sort]

	err = r.db.SelectContext(ctx, &result, rawQuery, args...)
	return
}

func (r *repository) GetMerchantTransactionDetail(ctx context.Context, merchantId, id string) (*walletTransactionModel.MerchantTransactionDetailResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/wallet/transaction/GetMerchantTransactionDetail")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
		at.uuid, at.type, IF(at.channel != '', at.channel, IFNULL(at.additional_info->>'$.referenceType', '')) AS channel, 
		at.created_at, at.updated_at,
		CASE
			WHEN at.debit > 0 THEN at.debit
			WHEN COALESCE(at.credit, 0) = 0
				AND NULLIF(at.additional_info->>'$.subPaymentSummary.totalCreditAmount','') IS NOT NULL
			THEN CAST(
				at.additional_info->>'$.subPaymentSummary.totalCreditAmount'
				AS DECIMAL(18,2)
			)
			ELSE COALESCE(at.credit, 0)
		END AS amount,
		CASE
			WHEN at.type = 'MERCHANT_PAYMENT' THEN cus.phone_number
			WHEN at.type = 'WITHDRAWAL' THEN w_user.name
			ELSE 'System'
		END AS created_by,
		CASE
			WHEN at.type = 'WITHDRAWAL' AND at.channel = 'BANK_TRANSFER'
				THEN JSON_OBJECT(
					'destination', at.channel, 'accountNumber', w.beneficiary_account_no, 'accountName', w.beneficiary_account_name, 'bankName', w.beneficiary_bank_name
				)
			WHEN at.type = 'WITHDRAWAL' AND at.channel = 'BALANCE_TRANSFER'
				THEN JSON_OBJECT('destination', at.channel, 'destinationAccountName', w.beneficiary_account_name)
			WHEN at.type = 'FEE'
				THEN JSON_OBJECT('linkedTransactionId', '-')
			WHEN at.type = 'MERCHANT_PAYMENT' 
				THEN JSON_OBJECT('username', cus.first_name)
			ELSE 
				JSON_OBJECT()
		END AS additional_info,
		at.status, IF(at.status != 'SUCCESS', at.status, IFNULL(at.settlement_status, 'SUCCESS')) AS settlement_status,
		IF(at.type IN ('MERCHANT_TOP_UP','WITHDRAWAL'), '-', IFNULL(at.merchant_reference_id, '')) AS merchant_reference_id
	FROM account_transactions at
	LEFT JOIN account_transactions at_cus ON at_cus.reference_id = at.reference_id AND at_cus.type = 'MERCHANT_PAYMENT' AND at_cus.account_id != at.account_id
	LEFT JOIN accounts acc_cus ON acc_cus.uuid = at_cus.account_id
	LEFT JOIN customers cus ON cus.uuid = acc_cus.reference_id
	LEFT JOIN withdrawals w ON w.id = at.reference_id AND at.type = 'WITHDRAWAL'
	LEFT JOIN users w_user ON w_user.uuid = w.created_by
	WHERE at.merchant_id = ? AND at.uuid = ?;`

	result := walletTransactionModel.MerchantTransactionDetailResp{}
	if err := r.db.GetContext(ctx, &result, rawQuery, merchantId, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_ = result.RawAdditionalInfo.Unmarshal(&result.AdditionalInfo)

	return &result, nil
}
