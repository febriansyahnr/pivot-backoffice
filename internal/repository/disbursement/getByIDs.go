package disbursementRepository

import (
	"context"
	"fmt"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) GetByIDs(ctx context.Context, ids []string) ([]*disbursementModel.Disbursement, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/FindDisbursementByIDs")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	if len(ids) == 0 {
		return []*disbursementModel.Disbursement{}, nil
	}

	var data []*disbursementModel.Disbursement

	query := `
		SELECT
			uuid,
		 	reference_id,
			merchant_id,
			bulk_id,
			purpose_id,
			sender_name,
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
			deleted_at
		FROM
			disbursements`

	idString := fmt.Sprintf("'%s'", ids[0])
	for _, id := range ids[1:] {
		idString += fmt.Sprintf(", '%s'", id)
	}
	query += fmt.Sprintf(" WHERE uuid IN (%s)", idString)

	if err := r.db.SelectContext(ctx, &data, query); err != nil {
		r.pdkLogger.Error(ctx, "error when get disbursement by uuid", logger.Error(err))
		return nil, err
	}

	return data, nil
}
