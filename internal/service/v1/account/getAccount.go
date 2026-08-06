package accountService

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (a *account) GetAccount(ctx context.Context, accountId uuid.UUID) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/GetAccount")
	defer segment.End()

	// Get Account By UUID
	account, err := a.accountRepo.GetByUUID(ctx, accountId)
	if err != nil {
		a.logger.Error(ctx, "error when get account by merchant id", logger.Error(err))
		return nil, err
	}

	now := time.Now().UTC()

	// Get All Transaction By Merchant and Account Between Last Update account and Now
	resp, err := a.accountTransactionRepo.GetAggregateTransactions(ctx, &orchestrator_model.GetAggregateRequest{
		MerchantID: account.ReferenceID,
		AccountID:  account.UUID,
		Statuses:   []string{constant.StatusSuccess},
		StartAt:    &account.LastUpdateBalanceAt,
		EndAt:      a.setYesterdayMidnight(ctx),
	})
	if err != nil {
		a.logger.Error(ctx, "error when calculate account by merchant, account and date", logger.Error(err))
		return nil, err
	}

	account.CurrentBalance = account.EODBalance + resp.SumOfDebit - resp.SumOfCredit
	account.LastUpdateBalanceAt = now

	return account, nil
}
