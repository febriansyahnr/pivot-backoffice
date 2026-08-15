package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) FindBulkDisbursementByID(ctx context.Context, id string) (*disbursementModel.BulkDisbursement, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindBulkDisbursementByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	var data disbursementModel.BulkDisbursement

	query := `
		SELECT
			uuid,
			merchant_id,
			file,
			file_failed,
			file_rejected,
			status,
			created_by,
			created_at,
			updated_at,
			deleted_at
		FROM
			bulk_disbursements
		WHERE uuid = ?`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.pdkLogger.Error(ctx, "bulk disbursement not found", logger.String("uuid", id))
			return nil, nil
		}

		r.pdkLogger.Error(ctx, "error when finding bulk disbursement data by id", logger.String("uuid", id), logger.Error(err))
		return &data, err
	}

	return &data, nil
}
