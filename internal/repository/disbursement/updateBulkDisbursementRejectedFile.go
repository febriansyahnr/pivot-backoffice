package disbursementRepository

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) UpdateBulkDisbursementRejectedFileByID(ctx context.Context, id, rejectedFilePath string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateBulkDisbursementRejectedFileByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	query := `
		UPDATE bulk_disbursements
		SET
		    file_rejected = ?, 
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		rejectedFilePath,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.pdkLogger.Error(ctx, "error when updating bulk disbursement", logger.Error(err))
		return err
	}

	return nil
}
