package transferRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *transferRepository) GetByID(ctx context.Context, id, merchantId string) (*transfer.Transfer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/GetById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var transfer transfer.Transfer
	query := `
			SELECT 
				t.uuid, 
				t.merchant_id, 
				t.recipient_id, 
				t.reference_id,
				t.currency, 
				t.amount, 
				t.status, 
				t.transfer_type,
				t.remarks,
				t.created_at, 
				t.updated_at, 
				t.deleted_at,
				n.name AS beneficiary
			FROM ` + tableName + `  t
			JOIN merchants n
			ON n.uuid = t.recipient_id
			WHERE t.merchant_id = ? AND t.uuid = ?
			LIMIT 1 `

	if err := r.db.GetContext(ctx, &transfer, query, merchantId, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding transfer by id", logger.Error(err), logger.Any("query", query), logger.Any("merchantId", merchantId), logger.Any("id", id))
		return nil, err
	}

	return &transfer, nil
}

// GetTransferTransaction retrieves the details of a transfer transaction based on the provided request.
// It queries the database to fetch the transaction details including sender and recipient information,
// transaction amount, type, fee details, payment reference, status, remarks, and creation timestamp.
func (r *transferRepository) GetTransferTransaction(ctx context.Context, req transfer.GetTransferTransactionRequest) (*transfer.TransferTransactionDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/GetTransferTransaction")
	defer segment.End()

	if req.MerchantID == "" || req.TransactionID == "" {
		return nil, errors.New("invalid request parameters")
	}

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var transferDetail transfer.TransferTransactionDetail
	query := `
			SELECT 
				t.uuid, 
				t.merchant_id AS sender_id, 
				m.name AS sender_name,
				t.recipient_id AS recipient_id, 
				n.name AS recipient_name,
				t.reference_id,
				t.currency, 
				CASE 
					WHEN t.recipient_id = ? THEN t.amount
					ELSE -t.amount
				END as amount,
				CASE 
					WHEN t.recipient_id = ? THEN '` + constant.TransferTypeIN + `'
					ELSE '` + constant.TransferTypeOUT + `'
				END as type,
				p.uuid AS payment_id,
				CASE 
					WHEN att.type = '` + constant.TypeFee + `' THEN att.debit
					ELSE 0
				END as fee_amount,
				att.currency as fee_currency,
				p.reference_id as payment_reference_id,
				t.status,
				t.remarks,
				t.created_at
			FROM ` + tableName + `  t
			LEFT JOIN payments p
			ON p.uuid = t.reference_id
			JOIN merchants m 
			ON m.uuid = t.merchant_id
			JOIN merchants n
			ON n.uuid = t.recipient_id
			LEFT JOIN account_transactions att
			ON att.reference_id = t.uuid AND (att.merchant_id = ? OR att.merchant_id = ?) AND att.type ='` + constant.TypeFee + `'
			WHERE (t.merchant_id = ? OR t.recipient_id = ?) AND t.uuid = ?
			LIMIT 1 `

	err := r.db.GetContext(ctx, &transferDetail, query, req.MerchantID, req.MerchantID, req.MerchantID, req.ParentID, req.MerchantID, req.MerchantID, req.TransactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.logger.Error(ctx, "error when finding transfer transaction by id", logger.Error(err), logger.String("merchantId", req.MerchantID), logger.String("id", req.TransactionID))
		return nil, err
	}

	return &transferDetail, nil
}
