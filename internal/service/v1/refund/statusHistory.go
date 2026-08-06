package refundService

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

func (s *RefundService) recordStatusHistory(ctx context.Context, req *statusHistoryModel.RecordRefundStatusHistoryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/refund/recordStatusHistory")
	defer segment.End()

	statusInfo, exists := constant.RefundStatusHistoryLabelsAndDescriptions[req.Status]
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
		s.logger.Error(ctx, "Failed to marshal refund status history metadata", logger.Error(err))
		return err
	}

	statusHistory := &statusHistoryModel.StatusHistory{
		ID:            uuid.New().String(),
		ReferenceType: constant.TypeRefund,
		ReferenceID:   req.RefundID,
		Status:        req.Status,
		Metadata:      types.NullJSONText{JSONText: metadataJSON, Valid: true},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.statusHistoriesRepo.Insert(ctx, statusHistory); err != nil {
		s.logger.Error(ctx, "Failed to insert refund status history", logger.Error(err), logger.Any("statusHistory", statusHistory))
		// Don't return error - status history is not critical for main flow
	}

	return nil
}

// Helper methods for common refund status history recording scenarios

func (s *RefundService) recordRefundPending(ctx context.Context, refundID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
		RefundID: refundID,
		Status:   constant.RefundStatusHistoryPending,
		Actor:    actor,
	})
}

func (s *RefundService) recordRefundWaitingBankTransfer(ctx context.Context, refundID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
		RefundID: refundID,
		Status:   constant.RefundStatusHistoryWaitingBankTransfer,
		Actor:    actor,
	})
}

func (s *RefundService) recordRefundSuccess(ctx context.Context, refundID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
		RefundID: refundID,
		Status:   constant.RefundStatusHistorySuccess,
		Actor:    actor,
	})
}

func (s *RefundService) recordRefundFailed(ctx context.Context, refundID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
		RefundID: refundID,
		Status:   constant.RefundStatusHistoryFailed,
		Actor:    actor,
	})
}

func (s *RefundService) recordRefundCancelled(ctx context.Context, refundID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
		RefundID: refundID,
		Status:   constant.RefundStatusHistoryCancelled,
		Actor:    actor,
	})
}
