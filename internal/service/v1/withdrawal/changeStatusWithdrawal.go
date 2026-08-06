package withdrawalService

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *withdrawalService) ChangeStatusWithdrawal(ctx context.Context, request *withdrawal.WithdrawalChangeStatusRequest) (*withdrawal.WithdrawalChangeStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/ChangeStatusWithdrawal")
	defer segment.End()

	// find account transactions by reference id
	wd, err := s.repo.FindById(ctx, request.WithdrawalID, request.MerchantID)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if wd == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	// Validate this is a bank transfer withdrawal (not balance transfer)
	if wd.Metadata.WithdrawType == constant.WithdrawalDestBalanceTransfer {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("retry is not supported for balance transfer withdrawals"))
	}

	// Get the account transaction for this withdrawal (contains the external ID for snap-core)
	accountTransaction, err := s.accountTrxRepo.FindByReference(ctx, request.WithdrawalID, constant.TypeWithdrawal)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if accountTransaction == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, errors.New("account transaction not found for this withdrawal"))
	}

	// Validate transaction status is already SUCCESS
	if accountTransaction.Status == constant.StatusSuccess {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrTransactionAlreadyInFinalStatus)
	}

	// update account transactions
	if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, accountTransaction.UUID.String(), request.Status, request.ReasonType, request.ReasonDescription); errUpdate != nil {
		s.logger.Error(ctx, "Update status account transaction", logger.Error(errUpdate))
		return nil, pkgErrs.New(response.HttpErrDatabase, errUpdate)
	}

	return &withdrawal.WithdrawalChangeStatusResponse{
		ID:         wd.Id,
		MerchantID: wd.MerchantId,
		Status:     request.Status,
	}, nil
}
