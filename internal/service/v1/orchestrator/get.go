package orchestrator_service

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *OrchestratorService) GetReferenceIdByTransactionIdAndType(ctx context.Context, transactionId, transactionType string) (referenceId string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetReferenceIdByTransactionIdAndType")
	defer segment.End()

	referenceId, err = s.accountTransactionRepo.GetReferenceIdByTransactionIdAndType(ctx, transactionId, transactionType)
	if err != nil {
		return "", pkgErrors.New(response.HttpErrDatabase, err)
	}

	return
}

func (s *OrchestratorService) FindByID(ctx context.Context, id string) (*orchestratorModel.AccountTransactionWithUseCase, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/orchestrator/FindByID")
	defer span.End()

	transaction, err := s.accountTransactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if transaction == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	return transaction, nil
}
