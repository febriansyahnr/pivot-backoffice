package unifiedPaymentService

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

func (s *UnifiedPaymentService) recordChargeStatusHistory(ctx context.Context, req *statusHistoryModel.RecordChargeStatusHistoryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/recordChargeStatusHistory")
	defer segment.End()

	statusInfo, exists := constant.ChargeStatusHistoryLabelsAndDescriptions[req.Status]
	if !exists {
		statusInfo = map[string]string{
			"label":       req.Status,
			"description": "Status updated",
		}
	}

	// For future enhancement, we can add reason type handling like disbursement
	description := statusInfo["description"]
	recommendation := statusInfo["recommendation"]

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
		s.logger.Error(ctx, "Failed to marshal charge status history metadata", logger.Error(err))
		return err
	}

	statusHistory := &statusHistoryModel.StatusHistory{
		ID:            uuid.New().String(),
		ReferenceType: constant.TypePayment,
		ReferenceID:   req.ChargeID,
		Status:        req.Status,
		Metadata:      types.NullJSONText{JSONText: metadataJSON, Valid: true},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.statusHistoriesRepo.Insert(ctx, statusHistory); err != nil {
		s.logger.Error(ctx, "Failed to insert charge status history", logger.Error(err), logger.Any("statusHistory", statusHistory))
		// Don't return error - status history is not critical for main flow
	}

	return nil
}

// Helper methods for common charge status history recording scenarios

func (s *UnifiedPaymentService) recordChargeWaitingForUserAction(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryWaitingForUserAction,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeWaitingForAuthentication(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryWaitingForAuthentication,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeWaitingForExternalFDS(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryWaitingForExternalFDS,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeProcessing(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryProcessing,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeWaitingForCapture(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryWaitingForCapture,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeSuccess(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistorySuccess,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeFailed(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryFailed,
		Actor:    actor,
	})
}

func (s *UnifiedPaymentService) recordChargeExpired(ctx context.Context, chargeID, actor string) {
	s.recordChargeStatusHistory(ctx, &statusHistoryModel.RecordChargeStatusHistoryRequest{
		ChargeID: chargeID,
		Status:   constant.ChargeStatusHistoryExpired,
		Actor:    actor,
	})
}
