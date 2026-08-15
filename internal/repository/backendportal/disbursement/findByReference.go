package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) FindByReference(ctx context.Context, referenceID string) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindByReference")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var data disbursementModel.DisbursementWithTransaction

	query := `SELECT 
		` + SelectDisbursementWithTransactionStr + `
	FROM disbursements d 
	LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id AND IFNULL(t.reason_type, '') != 'REVERSAL'
	LEFT JOIN users c ON c.uuid = d.created_by
	LEFT JOIN users a ON a.uuid = d.approved_by
	WHERE d.reference_id = ?`

	if err := r.db.GetContext(ctx, &data, query, referenceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.pdkLogger.Error(ctx, "disbursement by reference not found", logger.String("reference", referenceID))
			return nil, nil
		}

		r.pdkLogger.Error(ctx, "error when finding disbursement data by reference", logger.String("reference", referenceID), logger.Error(err))
		return &data, err
	}

	return &data, nil
}