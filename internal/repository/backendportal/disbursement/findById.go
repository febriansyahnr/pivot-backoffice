package disbursementRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) FindByID(ctx context.Context, id string) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var data disbursementModel.DisbursementWithTransaction

	query := `SELECT 
			` + SelectDisbursementWithTransactionStr + `
		FROM disbursements d 
		LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id AND IFNULL(t.reason_type, '') != 'REVERSAL'
		LEFT JOIN users c ON c.uuid = d.created_by
		LEFT JOIN users a ON a.uuid = d.approved_by
		WHERE d.uuid = ? AND d.type IN ('LOCAL_PAYOUT', 'INTERNATIONAL_PAYOUT')`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.pdkLogger.Error(ctx, "disbursement not found", logger.String("uuid", id))
			return nil, nil
		}

		r.pdkLogger.Error(ctx, "error when finding disbursement data by id", logger.String("uuid", id), logger.Error(err))
		return &data, err
	}
	if data.Metadata.Valid {
		_ = json.Unmarshal(data.Metadata.JSONText, &data.MetadataObj)
	}
	return &data, nil
}

func (r *DisbursementRepository) FindForReversalDisbursementById(ctx context.Context, merchantId, id string) (result *disbursementModel.DisbursementForReversal, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindForReversalDisbursementById")
	defer segment.End()

	result = &disbursementModel.DisbursementForReversal{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements,account_transactions")

	rawQuery := `SELECT
			uuid AS id, status, merchant_id, currency, amount, d.fee AS fee_amount, total_amount, reason_type,
			IFNULL((SELECT JSON_OBJECT('id', uuid, 'amount', debit, 'status', status, 'metadata', JSON_OBJECT()) FROM account_transactions WHERE reference_id = d.uuid AND type = 'DISBURSEMENT' ORDER BY updated_at DESC LIMIT 1), JSON_OBJECT()) AS transaction,
			IFNULL((SELECT JSON_OBJECT('id', uuid, 'amount', debit, 'status', status, 'metadata', IFNULL(additional_info, JSON_OBJECT())) FROM account_transactions WHERE reference_id = d.uuid AND type = 'FEE'), JSON_OBJECT()) AS fee
		FROM disbursements d WHERE uuid = ? AND merchant_id = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, result, rawQuery, id, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_ = result.RawFee.Unmarshal(&result.Fee)
	_ = result.RawTransaction.Unmarshal(&result.Transaction)

	return
}

// GetMerchantIDsForPayoutCallback retrieves the list of merchant IDs that should receive payout status callbacks.
// The data source for these merchant IDs is the bulk disbursement record.
func (r *DisbursementRepository) GetMerchantIDsForPayoutCallback(ctx context.Context, bulkId string) (merchantIds []string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetMerchantIDsForPayoutCallback")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "bulk_disbursements,merchants")

	rawQuery := `SELECT
		IFNULL(m.uuid, b.merchant_id) AS merchant_id, IFNULL(m.parent_id, '') AS parent_merchant_id
	FROM bulk_disbursements b
	LEFT JOIN merchants m ON m.uuid = b.created_by
	WHERE b.uuid = ?;`

	result := disbursementModel.MerchantIDForPayoutCallback{}
	if err = r.db.GetContext(ctx, &result, rawQuery, bulkId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.pdkLogger.Error(ctx, "Failed while getting merchant ids for payout callback", logger.Error(err))
		return nil, err
	}

	merchantIds = append(merchantIds, result.MerchantId)
	if result.ParentMerchantId != "" {
		merchantIds = append(merchantIds, result.ParentMerchantId)
	}
	return
}
