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

func (s *adjustmentService) ReleaseHoldedMerchantBalance(ctx context.Context, req *adjustModel.HoldMerchantBalanceRequest) (*adjustModel.HoldMerchantBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/ReleaseHoldedMerchantBalance")
	defer segment.End()

	// find account transaction by reference using req.UUID
	merchantId, _ := uuid.Parse(req.MerchantId)
	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantId, req.AccountType)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if account == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrAccountNotFound)
	}

	if account.HoldedBalance < req.Amount {
		req.Amount = account.HoldedBalance
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

	// update account transaction status to FAILED
	currentHoldedBalance := account.HoldedBalance
	account.HoldedBalance = currentHoldedBalance - req.Amount
	if err := s.accountRepo.UpdateHoldedBalance(ctx, account); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	adjustmentAmount := req.Amount

	if adjustmentAmount == 0 {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidAmount)
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	return &adjustModel.HoldMerchantBalanceResponse{
		Type:        string(constant.HoldedBalanceTypeRelease),
		Amount:      adjustmentAmount,
		MerchantID:  req.MerchantId,
		AccountType: req.AccountType,
	}, nil
}
