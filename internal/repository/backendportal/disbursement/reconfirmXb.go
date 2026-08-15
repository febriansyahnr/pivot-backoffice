package disbursementRepository

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
)

func (r *DisbursementRepository) ReconfirmXB(ctx context.Context, request *disbursementModel.ReconfirmXBRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/ReconfirmXB")
	defer segment.End()

	disbursementStatus, disbursementReasonType, _ := constant.MapXbProcessorStatusToCoreStatus(request.XBStatus)
	reasonDescription := constant.MapXbReasonTypeToDesc(disbursementReasonType)

	updateSqlQuery := `
		UPDATE disbursements
		SET
			status = :status,
			reason_type = :reason_type,
			reason_description = :reason_description,
			metadata = JSON_SET(metadata, '$.xbDetail.expiredAt', :expired_at),
			updated_at = :updated_at
		WHERE
			uuid = :uuid
	`

	_, err := r.db.NamedExecContext(ctx, updateSqlQuery, map[string]interface{}{
		"uuid":               request.PayoutId,
		"status":             disbursementStatus,
		"reason_type":        disbursementReasonType,
		"reason_description": reasonDescription,
		"expired_at":         request.ExtendedTime.Format(time.RFC3339),
		"updated_at":         time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}
