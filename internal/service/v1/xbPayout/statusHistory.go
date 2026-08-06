package xbPayoutService

import (
	"context"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) RecordStatusHistory(ctx context.Context, req *statusHistoryModel.RecordDisbursementStatusHistoryRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/RecordStatusHistory")
	defer segment.End()

	if req == nil {
		return nil
	}

	now := time.Now().UTC()

	metadata := statusHistoryModel.StatusHistoryMetadata{
		Actor: req.Actor,
	}

	switch req.Status {
	case constant.XbStatusCreated:
		metadata.Label = constant.StatusLabelPayoutCreated
		metadata.Description = constant.XbDisbursementReasonDescWaitingForConfirmation + "."

	case constant.XbStatusConfirmed:
		metadata.Label = constant.StatusLabelPayoutCreated
		metadata.Description = "Payout request confirmed."

	case constant.XbStatusInfoRequested:
		metadata.Label = "Information Requested"
		metadata.Description = "Further information requested by bank partner."
		metadata.Recommendation = "Please submit supporting information to Helpdesk."

	case constant.XbStatusComplianceVerification, constant.XbStatusComplianceApproved:
		metadata.Label, req.Status = "Information In Review", constant.XbDisbursementReasonTypeInReview
		metadata.Description = "Information submitted and in review by bank partner."

	case constant.XbStatusInProcess, constant.XbStatusPGProcessing, constant.XbStatusSentToBank, constant.XbStatusPending, constant.XbStatusRemindRecipient:
		metadata.Label, req.Status = "Processing", constant.XbDisbursementReasonTypeProcessing
		metadata.Description = "Transaction is being processed by our bank partner."

	case constant.XbStatusRejected:
		metadata.Label = "Rejected"
		metadata.Description = "Transaction rejected by beneficiary."

	case constant.XbStatusComplianceRejected:
		metadata.Label = "Rejected"
		metadata.Description = "Transaction rejected by compliance."

	case constant.XbStatusError:
		metadata.Label, req.Status = "Failed", constant.XbDisbursementReasonTypeFailed
		metadata.Description = "Transaction failed due to error from provider."

	case constant.XbStatusHttpError:
		metadata.Label, req.Status = "Error", constant.XbDisbursementReasonTypeError
		metadata.Description = "HTTP confirmation error. Transaction status unknown."

	case constant.XbStatusPaid:
		metadata.Label = "Success"
		metadata.Description = "Transaction has been successfully completed."

	case constant.XbStatusReturned:
		metadata.Label, req.Status = "Refunded", constant.XbDisbursementReasonTypeRefunded
		metadata.Description = "Transaction has been refunded."

	case constant.XbStatusCanceled:
		metadata.Label = "Canceled"
		metadata.Description = "Transaction has been cancelled."

	case constant.XbStatusExpired:
		metadata.Label = "Expired"
		metadata.Description = "Transaction has been expired."

	default:
		metadata.Label = util.ToTitle(req.Status)
		metadata.Description = "Transaction status changed."
	}

	statusHistory := &statusHistoryModel.StatusHistory{
		ID:            util.GenerateUUID().String(),
		ReferenceType: constant.TypeXB,
		ReferenceID:   req.DisbursementID,
		Status:        req.Status,
		Metadata:      types.NullJSONText{Valid: true},
		CreatedAt:     now, UpdatedAt: now,
	}
	statusHistory.Metadata.JSONText, _ = json.Marshal(metadata)

	if err = s.statusHistoriesRepo.Insert(ctx, statusHistory); err != nil {
		s.logger.Error(ctx, "Failed to insert status history", logger.Error(err), logger.Any("statusHistory", statusHistory))
		return err
	}
	return nil
}
