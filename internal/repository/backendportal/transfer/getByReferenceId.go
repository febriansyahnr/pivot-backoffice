package transferRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *transferRepository) GetByReferenceID(ctx context.Context, merchantId, recipientId, referenceId string) (*transfer.Transfer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/GetByReferenceId")
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
				t.deleted_at
			FROM ` + tableName + `  t
			WHERE t.merchant_id = ? AND t.recipient_id = ? AND t.reference_id = ?
			LIMIT 1 `

	if err := r.db.GetContext(ctx, &transfer, query, merchantId, recipientId, referenceId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding transfer by referenceId", logger.Error(err), logger.Any("query", query), logger.Any("merchantId", merchantId), logger.Any("referenceId", referenceId))
		return nil, err
	}

	return &transfer, nil
}
