package transferRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *transferRepository) Create(ctx context.Context, transfer *transfer.Transfer) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO ` + tableName + ` 
		(
			uuid, 
			reference_id, 
			merchant_id, 
			recipient_id, 
			currency, 
			amount, 
			status, 
			transfer_type,
			created_at, 
			updated_at, 
			deleted_at
		)
		VALUES (
			:uuid, 
			:reference_id, 
			:merchant_id, 
			:recipient_id, 
			:currency, 
			:amount, 
			:status, 
			:transfer_type,
			:created_at, 
			:updated_at, 
			:deleted_at
		)
	`
	affected, err := r.db.NamedExecContext(ctx, query, transfer)
	if err != nil {
		r.logger.Error(ctx, "error when inserting transfers data", logger.Error(err))
		return err
	}
	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting transfers data", logger.Error(err))
		return err
	}

	return nil

}
