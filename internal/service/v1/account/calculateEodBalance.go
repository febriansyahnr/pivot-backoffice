package accountService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (b *account) setYesterdayMidnight(ctx context.Context) *time.Time {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/account/setYesterdayMidnight")
	defer segment.End()

	// Get the current time
	now := time.Now().UTC()

	// Subtract one day from the current time
	yesterday := now.AddDate(0, 0, -1)

	// Set the time to 23:59
	yesterdayAt2359 := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, constant.LastNanoSecond, yesterday.Location())

	return &yesterdayAt2359
}

func (b *account) CalculateAccountEodBalance(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/CalculateAccountEodBalance")
	defer segment.End()

	// Get All Accounts
	accounts, err := b.accountRepo.FindAll(ctx)
	if err != nil {
		b.logger.Error(ctx, "error when get all accounts", logger.Error(err))
		return err
	}

	// Looping Accounts
	for _, account := range accounts {
		beginBalance := account.EODBalance
		endTimeAt := b.setYesterdayMidnight(ctx)

		// Calculate All Transaction by Merchant And account Between Last Update account and Now
		aggregateRequest := &orchestrator_model.GetAggregateRequest{
			MerchantID: account.ReferenceID,
			AccountID:  account.UUID,
			Statuses:   []string{constant.StatusSuccess},
			StartAt:    &account.LastUpdateBalanceAt,
			EndAt:      endTimeAt,
		}
		if account.UserType == constant.UserTypeCustomer {
			_, err := b.customerSvc.FindCustomerByID(ctx, account.ReferenceID.String())
			if err != nil {
				b.logger.Error(ctx, "error when get customer by id", logger.Error(err), logger.String("accountId", account.UUID.String()), logger.String("referenceId", account.ReferenceID.String()))
				continue
			}
			aggregateRequest.MerchantID = uuid.Nil
		}
		resp, err := b.accountTransactionRepo.GetAggregateTransactions(ctx, aggregateRequest)
		if err != nil {
			// TODO: Notify to slack
			b.logger.Error(ctx, "error when calculate account balance by merchant, balance and date", logger.Error(err))
			return err
		}

		b.logger.Info(
			ctx, "Balance summary for reference account "+account.ReferenceID.String(),
			logger.String("source", "CalculateAccountEodBalance"),
			logger.String("type", account.Name), logger.String("currency", account.Currency),
			logger.Float64("beginningBalance", beginBalance), logger.Float64("totalDebit", resp.SumOfDebit),
			logger.Float64("totalCredit", resp.SumOfCredit), logger.Float64("endingBalance", account.EODBalance),
			logger.String("lastUpdateBalanceAt", account.LastUpdateBalanceAt.String()),
		)

		var pendingTrxList []string
		if (resp.SumOfDebit + resp.SumOfCredit) > 0 {
			pendingTrxList, err = b.accountTransactionRepo.GetListOfTransactionReferenceIdsWithPendingStatus(
				ctx, aggregateRequest.MerchantID.String(), account.UUID.String(), account.LastUpdateBalanceAt, *endTimeAt,
			)
			if err != nil {
				b.logger.Error(ctx, "error when get list of transaction reference ids with pending status", logger.Error(err))
				return err
			}
		}
		if account.RequiresPendingBalanceCalculation() {
			aggregateRequest.StartAt = util.ValueToPtr(account.GetPendingTransactionCutoffOrBackdate())
			aggregateRequest.Statuses = []string{}
			aggregateRequest.PendingSettlementBalance = true

			earliestUpdatedAt, err := b.accountTransactionRepo.GetEarliestUpdatedAt(ctx, aggregateRequest)
			if err != nil {
				b.logger.Error(ctx, "error when get earliest updated at pending transaction", logger.Error(err))
				return err
			}
			if earliestUpdatedAt.IsZero() {
				earliestUpdatedAt = *b.setYesterdayMidnight(ctx)
			}
			account.PendingTransactionCutoffAt = &earliestUpdatedAt
		}
		ctxTrx, err := b.accountTransactionRepo.BeginTransaction(ctx)
		if err != nil {
			b.logger.Error(ctx, "failed while starting transaction session", logger.Error(err))
			return err
		}
		rollbackTransaction := func() {
			if err := b.accountTransactionRepo.RollbackTransaction(ctxTrx); err != nil {
				b.logger.Error(ctx, "failed to cancel transaction session", logger.Error(err))
			}
		}

		if resp.SumOfDebit > 0 || resp.SumOfCredit > 0 {
			account.EODBalance += resp.SumOfCredit - resp.SumOfDebit
			account.LastUpdateBalanceAt = *b.setYesterdayMidnight(ctx)

			if err := b.accountRepo.UpdateAccount(ctx, account); err != nil {
				rollbackTransaction()

				// TODO: Notify to slack
				b.logger.Error(ctx, "error when update account", logger.Error(err))
				return err
			}

			if err := b.accountTransactionRepo.RearrangeUpdatedAtForTransactionWithPendingStatus(ctxTrx, pendingTrxList, time.Now().UTC()); err != nil {
				rollbackTransaction()

				b.logger.Error(ctx, "error when rearrange updated at for transaction with pending status", logger.Error(err))
				return err
			}
		}

		// Insert daily account transaction
		date := *b.setYesterdayMidnight(ctx)
		uuidV7, _ := uuid.NewV7()
		if err := b.dailyAccountTransactionRepo.Upsert(ctx, &dailyAccountTransactionModel.DailyAccountTransaction{
			ID:           uuidV7.String(),
			AccountID:    account.UUID.String(),
			Date:         date,
			Timezone:     date.Location().String(),
			BegBalance:   beginBalance,
			DebitTrx:     resp.CountOfDebit,
			DebitAmount:  resp.SumOfDebit,
			CreditTrx:    resp.CountOfCredit,
			CreditAmount: resp.SumOfCredit,
			EODBalance:   account.EODBalance,
		}); err != nil {
			rollbackTransaction()

			b.logger.Error(ctx, "error when upsert daily account transaction", logger.Error(err))
			return err
		}

		if err := b.accountTransactionRepo.CommitTransaction(ctxTrx); err != nil {
			rollbackTransaction()

			b.logger.Error(ctx, "failed while performing transaction", logger.Error(err))
			return err
		}
	}

	return nil
}
