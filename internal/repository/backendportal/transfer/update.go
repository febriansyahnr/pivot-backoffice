package transferRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *transferRepository) Update(ctx context.Context, transfer *transfer.Transfer) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE ` + tableName + `  
		SET 
			status = ?, 
			reason_description = ?,
			updated_at = CURRENT_TIMESTAMP()
		WHERE merchant_id = ? AND uuid = ?
	`
	_, err := r.db.ExecContext(ctx, query, transfer.Status, transfer.ReasonDescription, transfer.MerchantID, transfer.UUID)
	if err != nil {
		r.logger.Error(ctx, "error when updating transfers data", logger.Error(err))
		return err
	}

	return nil

}
