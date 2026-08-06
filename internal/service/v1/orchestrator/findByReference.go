package orchestrator_service

import (
	"context"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *OrchestratorService) FindByReference(
	ctx context.Context,
	referenceID, referenceType string,
) (*orchestratorModel.AccountTransactionWithUseCase, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/FindByReference")
	defer segment.End()

	transaction, err := s.accountTransactionRepo.FindByReference(ctx, referenceID, referenceType)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	return transaction, nil
}
