package paymentService

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

func (s *PaymentService) recordStatusHistory(ctx context.Context, req *statusHistoryModel.RecordPaymentStatusHistoryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/recordStatusHistory")
	defer segment.End()

	statusInfo, exists := constant.PaymentStatusHistoryLabelsAndDescriptions[req.Status]
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
		s.logger.Error(ctx, "Failed to marshal status history metadata", logger.Error(err))
		return err
	}

	statusHistory := &statusHistoryModel.StatusHistory{
		ID:            uuid.New().String(),
		ReferenceType: constant.TypePayment,
		ReferenceID:   req.PaymentID,
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

func (s *PaymentService) recordPaymentPending(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryPending,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentRequirePaymentMethod(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryRequirePaymentMethod,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentRequireConfirmation(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryRequireConfirmation,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentRequireAction(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryRequireAction,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentProcessing(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryProcessing,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentSuccess(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistorySuccess,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentPaid(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryPaid,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentVoid(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryVoid,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentExpired(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryExpired,
		Actor:     actor,
	})
}

func (s *PaymentService) recordPaymentCancelled(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryCancelled,
		Actor:     actor,
	})
}

func (s *PaymentService) recordInvestigationInProcess(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryInvestigationInProcess,
		Actor:     actor,
	})
}

func (s *PaymentService) recordInvestigationSuccess(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryInvestigationSuccess,
		Actor:     actor,
	})
}

func (s *PaymentService) recordInvestigationFailed(ctx context.Context, paymentID, actor string) {
	s.recordStatusHistory(ctx, &statusHistoryModel.RecordPaymentStatusHistoryRequest{
		PaymentID: paymentID,
		Status:    constant.PaymentStatusHistoryInvestigationFailed,
		Actor:     actor,
	})
}
