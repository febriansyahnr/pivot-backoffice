package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *DisbursementRepository) FindByMerchantAndReference(ctx context.Context, merchantID, referenceID string) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindByMerchantAndReference")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var data disbursementModel.DisbursementWithTransaction

	query := `SELECT 
			` + SelectDisbursementWithTransactionStr + `
		FROM disbursements d 
		LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id AND IFNULL(t.reason_type, '') != 'REVERSAL'
		LEFT JOIN users c ON c.uuid = d.created_by
		LEFT JOIN users a ON a.uuid = d.approved_by
		WHERE d.merchant_id = ? AND d.reference_id = ?`

	if err := r.db.GetContext(ctx, &data, query, merchantID, referenceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.pdkLogger.Error(ctx, "disbursement by merchant and reference not found", logger.String("merchant", merchantID), logger.String("reference", referenceID))
			return nil, nil
		}

		r.pdkLogger.Error(ctx, "error when finding disbursement data by merchant and reference not found", logger.String("merchant", merchantID), logger.String("reference", referenceID), logger.Error(err))
		return &data, err
	}

	return &data, nil
}
