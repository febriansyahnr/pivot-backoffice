package orchestrator_service

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func (s *OrchestratorService) GetList(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetList")
	defer segment.End()

	list, err := s.accountTransactionRepo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
