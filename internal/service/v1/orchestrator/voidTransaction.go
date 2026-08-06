package orchestrator_service

import (
	"context"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func (s *OrchestratorService) VoidCreditcardTransaction(ctx context.Context, req *orchestratorModel.VoidTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/VoidCreditcardTransaction")
	defer segment.End()

	return s.accountTransactionRepo.VoidTransaction(ctx, req)
}
