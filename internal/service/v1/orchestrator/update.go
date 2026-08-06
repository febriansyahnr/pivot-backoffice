package orchestrator_service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx/types"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *OrchestratorService) UpdateStatusAccountTransaction(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateStatusAccountTransaction")
	defer segment.End()

	return s.accountTransactionRepo.UpdateStatusAccountTransaction(ctx, id, status, reasonType, reasonDescription)
}

func (s *OrchestratorService) UpdateStatusAccountTransactionByReferenceID(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateStatusAccountTransactionByReferenceID")
	defer segment.End()

	return s.accountTransactionRepo.UpdateStatusAccountTransactionByReferenceID(ctx, id, status, reasonType, reasonDescription)
}

func (s *OrchestratorService) UpdateReasonOnly(ctx context.Context, id string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateReasonOnly")
	defer segment.End()

	return s.accountTransactionRepo.UpdateReasonOnly(ctx, id, reasonType, reasonDescription)
}

func (s *OrchestratorService) UpdateAdditionalInfoByID(ctx context.Context, id string, additionalInfo []byte) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateAdditionalInfoByID")
	defer segment.End()

	return s.accountTransactionRepo.UpdateAdditionalInfoByID(ctx, id, types.NullJSONText{
		Valid:    true,
		JSONText: additionalInfo,
	})
}

func (s *OrchestratorService) UpdateStatusAccountAndAdditionalInfoTransaction(ctx context.Context, id string, status string, reasonType string, reasonDescription string, additionalInfo []byte) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateStatusAccountAndAdditionalInfoTransaction")
	defer segment.End()

	return s.accountTransactionRepo.UpdateTransactionsStatusAndAdditionalInfoByID(ctx, id,
		status, reasonType, reasonDescription,
		types.NullJSONText{
			Valid:    true,
			JSONText: additionalInfo,
		})
}

func (s *OrchestratorService) UpdateProcessorAndReconReferenceByID(ctx context.Context, id string, processorReferenceName, processorReferenceId, reconReference string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateProcessorAndReconReferenceByID")
	defer segment.End()

	return s.accountTransactionRepo.UpdateProcessorAndReconReference(ctx, id, processorReferenceName, processorReferenceId, reconReference)
}

func (s *OrchestratorService) UpdateTransactionTimestamp(ctx context.Context, id string, transactionTimestamp time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateTransactionTimestamp")
	defer segment.End()

	return s.accountTransactionRepo.UpdateTransactionTimestamp(ctx, id, transactionTimestamp)
}

func (s *OrchestratorService) UpdateTransaction(ctx context.Context, request *orchestratorModel.UpdateTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/UpdateTransaction")
	defer segment.End()

	if err := s.accountTransactionRepo.UpdateTransactionDetail(ctx, *request); err != nil {
		s.logger.Error(ctx, "error when update transaction", logger.Error(err), logger.String("transactionId", request.TransactionID))
		return err
	}
	return nil
}
