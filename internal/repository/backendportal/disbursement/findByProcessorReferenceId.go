package disbursementRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) FindByProcessorReferenceID(ctx context.Context, processorReferenceID string) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindByProcessorReferenceID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var data disbursementModel.DisbursementWithTransaction

	query := `SELECT 
			` + SelectDisbursementWithTransactionStr + `
		FROM disbursements d 
		LEFT JOIN account_transactions t ON t.type = '` + constant.TypeDisbursement + `' AND d.uuid = t.reference_id AND IFNULL(t.reason_type, '') != 'REVERSAL'
		LEFT JOIN users c ON c.uuid = d.created_by
		LEFT JOIN users a ON a.uuid = d.approved_by
		WHERE d.processor_reference_id = ?`

	if err := r.db.GetContext(ctx, &data, query, processorReferenceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.pdkLogger.Warn(ctx, "disbursement not found", logger.String("processorReferenceID", processorReferenceID))
			return nil, nil
		}

		r.pdkLogger.Error(ctx, "error when finding disbursement data by id", logger.String("processorReferenceID", processorReferenceID), logger.Error(err))
		return &data, err
	}
	if data.Metadata.Valid {
		_ = json.Unmarshal(data.Metadata.JSONText, &data.MetadataObj)
	}
	return &data, nil
}
