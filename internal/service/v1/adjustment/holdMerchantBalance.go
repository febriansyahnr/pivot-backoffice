package adjustment

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *adjustmentService) HoldMerchantBalance(ctx context.Context, req *adjustModel.HoldMerchantBalanceRequest) (*adjustModel.HoldMerchantBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/HoldMerchantBalance")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, req.MerchantId)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	availableBalance, err := s.orchestrator.GetAvailableMerchantBalance(ctx, req.MerchantId, req.AccountType)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if req.Amount > availableBalance {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrInsufficientBalance)
	}

	merchantID, _ := uuid.Parse(req.MerchantId)

	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantID, req.AccountType)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if account == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrAccountNotFound)
	}

	if ctx, err = s.repo.BeginTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.repo.RollbackTransaction(ctx); e != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(e))
			}
		}
	}()

	adjustmentAmount := req.Amount * -1

	if adjustmentAmount == 0 {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidAmount)
	}

	currentHoldedBalance := account.HoldedBalance
	account.HoldedBalance = currentHoldedBalance + req.Amount
	if err = s.accountRepo.UpdateHoldedBalance(ctx, account); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	return &adjustModel.HoldMerchantBalanceResponse{
		Type:        string(constant.HoldedBalanceTypeHold),
		Amount:      req.Amount,
		MerchantID:  req.MerchantId,
		AccountType: req.AccountType,
	}, nil
}
