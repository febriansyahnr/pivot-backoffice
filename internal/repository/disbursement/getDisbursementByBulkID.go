package disbursementRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) GetAllDisbursementByBulkID(ctx context.Context, bulkID string) ([]*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/disbursement/GetAllDisbursementByBulkID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	data := make([]*disbursementModel.DisbursementWithTransaction, 0)

	query := `SELECT 
    	` + SelectDisbursementWithTransactionStr + `
	FROM disbursements d
	LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id AND IFNULL(t.reason_type, '') != 'REVERSAL'
	LEFT JOIN users c ON c.uuid = d.created_by
	LEFT JOIN users a ON a.uuid = d.approved_by
	WHERE bulk_id = ?`

	err := r.db.SelectContext(ctx, &data, query, bulkID)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when get disbursement by bulk_id", logger.Error(err), logger.String("bulkId", bulkID))
		return nil, err
	}

	return data, nil
}
