package orchestrator_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *OrchestratorService) GetAvailableMerchantBalance(ctx context.Context, merchantID, balanceName string) (float64, error) {
	// Parse merchant ID
	merchantUUID, err := uuid.Parse(merchantID)
	if err != nil {
		s.logger.Error(ctx, "error parsing merchant id", logger.Error(err))
		return 0, pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	// Find disbursement balance by merchantID
	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantUUID, balanceName)
	if err != nil {
		return 0, err
	}
	if account == nil {
		account, err = account_model.NewAccount(&account_model.NewAccountRequest{
			ReferenceID: merchantUUID,
			Usecase:     balanceName,
			Currency:    "IDR",
			UserType:    constant.UserTypeMerchant,
		})
		if err != nil {
			return 0, pkgErrors.New(httpResponse.HttpErrRequest, err)
		}

		if err = s.accountRepo.Create(ctx, account); err != nil {
			return 0, err
		}
	}

	// Calculate not settled yet balance
	startAt := account.LastUpdateBalanceAt
	endAt := time.Now().UTC()
	aggregateTransactionRequest := &orchestrator_model.GetAggregateRequest{
		MerchantID: merchantUUID,
		AccountID:  account.UUID,
		Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
		StartAt:    &startAt,
		EndAt:      &endAt,
	}
	aggregateTransaction, err := s.accountTransactionRepo.GetAggregateTransactions(ctx, aggregateTransactionRequest)
	if err != nil {
		return 0, err
	}

	availableBalance := account.EODBalance + aggregateTransaction.SumOfCredit - aggregateTransaction.SumOfDebit - account.HoldedBalance

	return availableBalance, nil
}

func (s *OrchestratorService) GetMerchantBulkBalances(ctx context.Context, request *account_model.GetBulkBalanceRequest) (map[string]*account_model.AvailableBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetMerchantBulkBalances")
	defer segment.End()

	accountList, err := s.accountSvc.GetMerchantAccounts(ctx, request.MerchantIDs, request.Usecase)
	if err != nil {
		s.logger.Error(ctx, "error when retrieve merchant accounts", logger.Error(err), logger.Any("merchantIds", request.MerchantIDs), logger.String("usecase", request.Usecase))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetBulkBalance)
	}
	if len(accountList) == 0 {
		return nil, nil
	}

	var (
		accountClauses = make([]orchestrator_model.AccountsAggregationClause, len(accountList))
		i              = 0
		endTimeAt      = time.Now().UTC().Add(time.Hour * 1)
	)
	for _, account := range accountList {
		accountClauses[i] = orchestrator_model.AccountsAggregationClause{
			AccountID: account.UUID.String(),
			StartAt:   &account.LastUpdateBalanceAt,
			EndAt:     &endTimeAt,
		}
		i++
	}
	aggregateResponse, err := s.accountTransactionRepo.GetBulkAggregateTransactions(ctx, &orchestrator_model.BulkGetAggregateRequest{
		IncludeFeeIndirectDeduction: false,
		AccountClauses:              accountClauses,
		Statuses:                    []string{constant.StatusPending, constant.StatusSuccess},
	})
	if err != nil {
		s.logger.Error(ctx, "error when calculate aggregate transaction")
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetBulkBalance)
	}
	accountAggregationMap := make(map[string]*orchestrator_model.BulkAggregateResponse, len(aggregateResponse))
	for _, aggregate := range aggregateResponse {
		accountAggregationMap[aggregate.AccountID] = aggregate
	}

	responseMap := make(map[string]*account_model.AvailableBalanceResponse, len(accountList))
	for merchantId, account := range accountList {
		aggregationDetail := accountAggregationMap[account.UUID.String()]
		balanceResponse := &account_model.AvailableBalanceResponse{
			Currency: constant.CurrencyIDR,
			Balance:  account.EODBalance,
		}
		if aggregationDetail != nil {
			balanceResponse.Balance += aggregationDetail.SumOfCredit - aggregationDetail.SumOfDebit
		}
		responseMap[merchantId.String()] = balanceResponse
	}

	return responseMap, nil
}
