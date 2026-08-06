package disbursementRepository

import (
	"context"
	"errors"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) Insert(ctx context.Context, request *disbursementModel.Disbursement) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/Insert")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	query := `
		INSERT INTO disbursements
		(
			uuid,
		 	reference_id,
			merchant_id,
			bulk_id,
			purpose_id,
			sender_name,
		 	account_inquiry_id,
			beneficiary_bank_code,
			beneficiary_bank_name,
			beneficiary_account_no,
			beneficiary_account_name,
			processor_reference_id,
			currency,
			amount,
			fee,
			total_amount,
			status,
			reason_type,
			reason_description,
		 	remark,
			created_from,
			created_by,
			approved_by,
			approved_at,
			created_at,
			updated_at,
			deleted_at, metadata
		) VALUES (
		    :uuid,
		    :reference_id,
			:merchant_id,
			:bulk_id,
			:purpose_id,
			:sender_name,
			:account_inquiry_id,
			:beneficiary_bank_code,
			:beneficiary_bank_name,
			:beneficiary_account_no,
			:beneficiary_account_name,
			:processor_reference_id,
			:currency,
			:amount,
			:fee,
			:total_amount,
			:status,
			:reason_type,
			:reason_description,
		    :remark,
			:created_from,
			:created_by,
			:approved_by,
			:approved_at,
			:created_at,
			:updated_at,
			:deleted_at, :metadata
		)`

	affected, err := r.db.NamedExecContext(ctx, query, request)
	if err != nil {
		r.pdkLogger.Error(ctx, "error when inserting disbursements", logger.Error(err))
		return err
	}

	if !affected {
		r.pdkLogger.Error(ctx, "failed when inserting disbursements", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
