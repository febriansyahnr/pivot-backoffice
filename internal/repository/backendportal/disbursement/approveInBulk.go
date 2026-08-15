package disbursementRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) ApproveInBulk(ctx context.Context, ids []string, approvedBy string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateStatusByIDs")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	if len(ids) == 0 {
		err := fmt.Errorf("no disbursements to update")
		r.pdkLogger.Error(ctx, err.Error(), logger.Error(err))
		return err
	}

	query := `
		UPDATE disbursements
		SET
		    status = ?, 
		    approved_by = ?,
		    approved_at = ?,
			updated_at = ?`

	idString := fmt.Sprintf("'%s'", ids[0])
	for _, id := range ids[1:] {
		idString += fmt.Sprintf(", '%s'", id)
	}
	query += fmt.Sprintf(" WHERE uuid IN (%s)", idString)
	query += fmt.Sprintf(" AND status = '%s'", constant.DisbursementStatusWaiting)

	affected, err := r.db.ExecContext(ctx, query, constant.DisbursementStatusApproved, approvedBy, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		r.pdkLogger.Error(ctx, "error when updating disbursement", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.pdkLogger.Error(ctx, err.Error(), logger.Error(err))
		return err
	}

	return nil
}
