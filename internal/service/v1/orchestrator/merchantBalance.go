package orchestrator_service

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (s *OrchestratorService) GetMerchantBalance(ctx context.Context, request model.GetMerchantBalanceRequest) (*model.GetMerchantBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetMerchantBalance")
	defer segment.End()

	merchantUUID, err := uuid.Parse(request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "Failed while parsing merchant ID to UUID format", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid)
	}

	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantUUID, request.BalanceName)
	if err != nil {
		s.logger.Error(ctx, "Failed to find merchant account by name", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	if account == nil {
		// Return the default value using IDR currency. Currently, merchant balance only supports IDR.
		return &model.GetMerchantBalanceResponse{
			AvailableBalance: commonModel.Amount{Currency: constant.CurrencyIDR, Value: "0.00"},
			PendingBalance:   commonModel.Amount{Currency: constant.CurrencyIDR, Value: "0.00"},
		}, nil
	}

	var (
		aggregateRequest = model.GetAggregateRequest{
			MerchantID: merchantUUID,
			AccountID:  account.UUID,
			Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
			StartAt:    &account.LastUpdateBalanceAt,
			EndAt:      &request.Date,
		}
		pendingBalanceRequest            = aggregateRequest
		pendingBalance, availableBalance = 0.0, 0.0
	)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		newCtx := segment.NewGoroutine(ctx)
		result, err := s.accountTransactionRepo.GetAggregateTransactions(newCtx, &aggregateRequest)
		if err != nil {
			s.logger.Error(newCtx, "Failed while get aggregate transactions", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
		}
		availableBalance = account.EODBalance + result.SumOfCredit - result.SumOfDebit - account.HoldedBalance

		if !account.RequiresPendingBalanceCalculation() {
			pendingBalance = result.SumOfPendCredit - result.SumOfPendDebit
		}
		return nil
	})
	group.Go(func() error {
		newCtx := segment.NewGoroutine(ctx)
		if !account.RequiresPendingBalanceCalculation() {
			return nil
		}

		pendingBalanceRequest.Statuses = []string{}
		pendingBalanceRequest.StartAt = util.ValueToPtr(account.GetPendingTransactionCutoffOrBackdate())
		pendingBalance, err = s.accountTransactionRepo.CalculatePendingBalance(newCtx, &pendingBalanceRequest)
		if err != nil {
			s.logger.Error(newCtx, "Failed while calculate pending balance", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}

	return &model.GetMerchantBalanceResponse{
		AvailableBalance: commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance).StringFixed(2),
		},
		PendingBalance: commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(pendingBalance).StringFixed(2),
		},
		TotalBalance: commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance + pendingBalance).StringFixed(2),
		},
	}, nil
}
