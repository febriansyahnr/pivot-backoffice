package disbursementRepository

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) Reject(ctx context.Context, id, reasonType, reasonDescription, rejectedBy string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateStatusAndReasonByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	query := `
		UPDATE disbursements
		SET
		    status = ?, 
			reason_type = ?,
			reason_description = ?,
			approved_by = ?,
			approved_at = ?,
			updated_at = ?
		WHERE uuid = ? AND status = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		constant.DisbursementStatusRejected,
		reasonType,
		reasonDescription,
		rejectedBy,
		time.Now().UTC(),
		time.Now().UTC(),
		id,
		constant.DisbursementStatusWaiting,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.pdkLogger.Error(ctx, "error when updating disbursement", logger.Error(err))
		return err
	}

	return nil
}
