package ledgerService

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *LedgerService) GetLedgerBalance(ctx context.Context, accountId uuid.UUID) (*ledger_model.LedgerBalance, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledgerService/GetLedgerBalance")
	defer segment.End()

	account, err := s.accountRepo.GetByUUID(ctx, accountId)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrFindAccount)
	}
	if account == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrNoRowsData)
	}

	endAt := time.Now().UTC()
	data, err := s.repo.GetAggregateTransactions(ctx, &orchestrator_model.GetAggregateRequest{
		AccountID: accountId,
		Statuses:  []string{constant.StatusSuccess, constant.StatusPending},
		StartAt:   &account.LastUpdateBalanceAt,
		EndAt:     &endAt,
	})
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetBalance)
	}

	balance := &ledger_model.LedgerBalance{
		Balance:  account.EODBalance + data.SumOfCredit - data.SumOfDebit,
		Currency: constant.CurrencyIDR,
	}
	return balance, nil
}

func (s *LedgerService) CalculateBulkLedgerBalance(ctx context.Context, request *account_model.CalculateBulkLedgerBalanceRequest) (*ledger_model.LedgerBalance, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledgerService/CalculateBulkLedgerBalance")
	defer segment.End()

	var (
		startTime    = time.Time{}
		totalBalance float64
	)

	accountList, err := s.accountRepo.GetByIDs(ctx, request.AccountIDs)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrFindAccount)
	}
	if len(accountList) == 0 {
		return &ledger_model.LedgerBalance{}, nil
	}
	for _, account := range accountList {
		if account.LastUpdateBalanceAt.After(startTime) {
			startTime = account.LastUpdateBalanceAt
		}
		totalBalance += account.EODBalance
	}

	endAt := time.Now().UTC()
	data, err := s.repo.GetAggregateTransactions(ctx, &orchestrator_model.GetAggregateRequest{
		AccountIDs: request.AccountIDs,
		Statuses:   []string{constant.StatusSuccess, constant.StatusPending},
		StartAt:    &startTime,
		EndAt:      &endAt,
	})
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetBulkBalance)
	}

	balance := &ledger_model.LedgerBalance{
		Balance:  totalBalance + data.SumOfCredit - data.SumOfDebit,
		Currency: constant.CurrencyIDR,
	}
	return balance, nil
}
