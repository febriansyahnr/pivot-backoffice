package orchestrator_service

import (
	"context"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *OrchestratorService) GetWalletCustomersTotalBalance(ctx context.Context, request *orchestratorModel.GetWalletTotalBalanceRequest) (float64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetWalletCustomersTotalBalance")
	defer segment.End()

	totalBalance, err := s.accountTransactionRepo.GetWalletCustomersTotalBalance(ctx, request)
	if err != nil {
		return 0, pkgErr.New(response.HttpErrInternal, err)
	}

	return totalBalance, nil
}
