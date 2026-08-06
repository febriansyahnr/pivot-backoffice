package withdrawalService

import (
	"context"
	"errors"
	"math"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *withdrawalService) Preparation(ctx context.Context, request *withdrawal.PreparationRequest) (*withdrawal.PreparationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/Preparation")
	defer segment.End()

	bankAccounts, err := s.bankAccountRepo.GetListBankAccount(ctx, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "Get list bank account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if bankAccounts == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found"))
	}

	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantId, request.AccountName)
	if err != nil {
		s.logger.Error(ctx, "Get available merchant balance", logger.Error(err))
		return nil, err
	}

	maxWithdrawableAmount := math.Floor(availableBalance)

	return &withdrawal.PreparationResponse{
		MerchantId:       request.MerchantId,
		AccountName:      request.AccountName,
		AvailableBalance: maxWithdrawableAmount,
		BankAccounts:     bankAccounts,
	}, nil
}
