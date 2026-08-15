package disbursementRepository

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) UpdateBulkDisbursementStatusByID(ctx context.Context, id, status string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateBulkDisbursementStatusByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	query := `
		UPDATE bulk_disbursements
		SET
		    status = ?, 
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
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
