package disbursementRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *DisbursementRepository) InsertBulkDisbursement(ctx context.Context, request *disbursementModel.BulkDisbursement) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/InsertBulkDisbursement")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	query := `
		INSERT INTO bulk_disbursements
		(
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
		) VALUES (
		    :uuid,
		 	:merchant_id,
			:file,
			:file_failed,
			:file_rejected,
			:status,
			:created_by,
			:created_at,
			:updated_at,
			:deleted_at
		)`

	affected, err := r.db.NamedExecContext(ctx, query, request)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when inserting bulk_disbursements", logger.Error(err))
		return err
	}

	if !affected {
		r.pdkLogger.Error(ctx, "failed when inserting bulk_disbursements", logger.Error(constant.ErrNoRowsAffected))
		return constant.ErrNoRowsAffected
	}

	return nil
}
