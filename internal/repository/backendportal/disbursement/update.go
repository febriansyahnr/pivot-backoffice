package disbursementRepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *DisbursementRepository) UpdateBankReferenceNo(ctx context.Context, id, bankReferenceNo string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateBankReferenceNo")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	query := `
		UPDATE disbursements
		SET
		    bank_reference_no = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		bankReferenceNo,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.pdkLogger.Error(ctx, "error when updating bank reference no", logger.Error(err))
		return err
	}

	return nil
}

func (r *DisbursementRepository) UpdateProcessorReferenceIdAndBankReferenceNo(ctx context.Context, id, processorReferenceId, bankReferenceNo string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateProcessorReferenceIdAndBankReferenceNo")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	query := `
		UPDATE disbursements
		SET
		    processor_reference_id = ?, 
		    bank_reference_no = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		processorReferenceId,
		bankReferenceNo,
		time.Now().UTC(),
		id,
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

func (r *DisbursementRepository) UpdateProcessorReferenceByID(ctx context.Context, request *disbursementModel.Disbursement) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateProcessorAndReconReferenceByID")
	defer span.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	rawQuery := `
		UPDATE disbursements
		SET processor_reference_id = :processor_reference_id,
			status = :status,
			updated_at = :updated_at
		WHERE uuid = :uuid
		`

	_, err := r.db.NamedExecContext(ctx, rawQuery, request)
	if err != nil {
		r.pdkLogger.Error(ctx, "UpdateProcessorAndReconReferenceByID | error when updating disbursement", logger.Error(err))
		return err
	}

	return nil
}

func (r *DisbursementRepository) UpdateReasonByIDs(ctx context.Context, ids []string, reasonType, reasonDescription string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateReasonByIDs")
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
		    reason_type = ?, 
		    reason_description = ?,
			updated_at = ?
		`

	idString := fmt.Sprintf("'%s'", ids[0])
	for _, id := range ids[1:] {
		idString += fmt.Sprintf(", '%s'", id)
	}

	query += fmt.Sprintf(" WHERE uuid IN (%s)", idString)
	_, err := r.db.ExecContext(
		ctx,
		query,
		reasonType,
		reasonDescription,
		time.Now().UTC(),
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.pdkLogger.Error(ctx, "error when updating disbursement", logger.Error(err), logger.Any("ids", ids))
		return err
	}

	return nil
}

func (r *DisbursementRepository) UpdateReversalTransaction(ctx context.Context, id, reasonType, reasonDescription, createdBy string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateReversalTransaction")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	rawQuery := `UPDATE disbursements 
			SET reason_type = ?, reason_description = ?, updated_at = ?, metadata = JSON_SET(metadata, '$.reversalBy', ?)
		WHERE uuid = ?;`
	_, err := r.db.ExecContext(ctx, rawQuery, reasonType, reasonDescription, time.Now().UTC(), createdBy, id)
	return err
}

func (r *DisbursementRepository) UpdateStatusAndReasonByID(ctx context.Context, id, status string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/UpdateStatusAndReasonByID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	rawQuery := `UPDATE disbursements 
			SET status = ?, reason_type = ?, reason_description = ?, updated_at = ?
		WHERE uuid = ?;`
	_, err := r.db.ExecContext(ctx, rawQuery, status, reasonType, reasonDescription, time.Now().UTC(), id)
	return err
}
