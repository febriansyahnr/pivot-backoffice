package disbursementService

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) recordStatusHistory(ctx context.Context, req *statusHistoryModel.RecordDisbursementStatusHistoryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/recordStatusHistory")
	defer segment.End()

	statusInfo, exists := constant.DisbursementStatusHistoryLabelsAndDescriptions[req.Status]
	if !exists {
		statusInfo = map[string]string{
			"label":       req.Status,
			"description": "Status updated",
		}
	}

	// For failed status, check if we have specific reason type descriptions
	description := statusInfo["description"]
	recommendation := statusInfo["recommendation"]

	if req.ReasonType != "" {
		if req.Status == constant.DisbursementStatusHistoryFailed {
			if reasonInfo, reasonExists := constant.DisbursementFailedReasonDescriptions[req.ReasonType]; reasonExists {
				description = reasonInfo["description"]
				recommendation = reasonInfo["recommendation"]
			}
		} else if req.Status == constant.DisbursementStatusHistoryProcessing {
			if reasonInfo, reasonExists := constant.DisbursementProcessingReasonDescriptions[req.ReasonType]; reasonExists {
				description = reasonInfo["description"]
				recommendation = reasonInfo["recommendation"]
			}
		}
	}

	now := time.Now().UTC()
	metadata := statusHistoryModel.StatusHistoryMetadata{
		Label:       statusInfo["label"],
		Description: description,
		Actor:       req.Actor,
	}

	// Set metadata fields if provided
	if recommendation != "" {
		metadata.Recommendation = recommendation
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		s.logger.Error(ctx, "Failed to marshal status history metadata", logger.Error(err))
		return err
	}

	statusHistory := &statusHistoryModel.StatusHistory{
		ID:            uuid.New().String(),
		ReferenceType: constant.TypeDisbursement,
		ReferenceID:   req.DisbursementID,
		Status:        req.Status,
		Metadata:      types.NullJSONText{JSONText: metadataJSON, Valid: true},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.statusHistoriesRepo.Insert(ctx, statusHistory); err != nil {
		s.logger.Error(ctx, "Failed to insert status history", logger.Error(err), logger.Any("statusHistory", statusHistory))
		// Don't return error - status history is not critical for main flow
	}

	return nil
}

// Helper methods for common status history recording scenarios

func (s *DisbursementService) recordDisbursementWaiting(ctx context.Context, disbursementID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryWaiting,
		Actor:          actor,
	})
}

func (s *DisbursementService) recordDisbursementApproved(ctx context.Context, disbursementID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryApproved,
		Actor:          actor,
	})
}

func (s *DisbursementService) recordDisbursementWaitingForTopUp(ctx context.Context, disbursementID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryWaitingForTopUp,
		Actor:          constant.UserSystemType,
	})
}

func (s *DisbursementService) recordDisbursementRejected(ctx context.Context, disbursementID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryRejected,
		Actor:          actor,
	})
}

func (s *DisbursementService) recordDisbursementProcessing(ctx context.Context, disbursementID, reasonType string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryProcessing,
		Actor:          constant.UserSystemType,
		ReasonType:     reasonType,
	})
}

func (s *DisbursementService) recordDisbursementSuccess(ctx context.Context, disbursementID string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistorySuccess,
		Actor:          constant.UserSystemType,
	})
}

func (s *DisbursementService) recordDisbursementFailed(ctx context.Context, disbursementID, reasonType string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryFailed,
		Actor:          constant.UserSystemType,
		ReasonType:     reasonType,
	})
}

func (s *DisbursementService) recordDisbursementCancelled(ctx context.Context, disbursementID, reasonType string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursementID,
		Status:         constant.DisbursementStatusHistoryCancelled,
		Actor:          constant.UserSystemType,
		ReasonType:     reasonType,
	})
}
