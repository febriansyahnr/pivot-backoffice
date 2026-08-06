package accountService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (b *account) CalculateDailyAccountTransaction(ctx context.Context, location *time.Location) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/CalculateAccountEodBalance")
	defer segment.End()

	// Get All Accounts
	accounts, err := b.accountRepo.FindAll(ctx)
	if err != nil {
		b.logger.Error(ctx, "error when get all accounts", pdkLogger.Error(err))
		return err
	}

	now := time.Now().In(location)
	yesterday := now.AddDate(0, 0, -1)
	locationStartAt := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 00, 00, 00, 0, yesterday.Location())
	locationEndAt := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, constant.LastNanoSecond, yesterday.Location())

	// Convert to UTC
	utcStartAt := locationStartAt.UTC()
	utcEndAt := locationEndAt.UTC()

	// Looping Accounts
	for _, account := range accounts {
		beginBalance := 0.0

		// get latest daily account transaction
		latestDailyAccountTransaction, err := b.dailyAccountTransactionRepo.FindLatestByAccountIDAndTimezone(ctx, account.UUID.String(), locationEndAt.Location().String())
		if err != nil {
			return err
		}

		if latestDailyAccountTransaction == nil {
			utcStartAt = account.CreatedAt
		} else {
			latestDailyTrxLocation, _ := time.LoadLocation(latestDailyAccountTransaction.Timezone)
			latestDailyTrxTimeInLocation := time.Date(
				latestDailyAccountTransaction.Date.Year(),
				latestDailyAccountTransaction.Date.Month(),
				latestDailyAccountTransaction.Date.Day(),
				0, 0, 0, 0,
				latestDailyTrxLocation,
			)

			utcStartAt = latestDailyTrxTimeInLocation.Add(24 * time.Hour).UTC()
			beginBalance = latestDailyAccountTransaction.EODBalance
		}

		// Calculate All Transaction by Merchant And account Between Last Update account and Now
		aggregateRequest := &orchestrator_model.GetAggregateRequest{
			MerchantID: account.ReferenceID,
			AccountID:  account.UUID,
			Statuses:   []string{constant.StatusSuccess},
			StartAt:    &utcStartAt,
			EndAt:      &utcEndAt,
		}
		if account.UserType == constant.UserTypeCustomer {
			customer, err := b.customerSvc.FindCustomerByID(ctx, account.ReferenceID.String())
			if err != nil {
				b.logger.Error(ctx, "error when get customer by id", logger.Error(err), logger.String("accountId", account.UUID.String()), logger.String("referenceId", account.ReferenceID.String()))
				continue
			}
			aggregateRequest.MerchantID = uuid.MustParse(customer.MerchantID)
		}
		resp, err := b.accountTransactionRepo.GetAggregateTransactions(ctx, aggregateRequest)
		if err != nil {
			b.logger.Error(ctx, "error when get aggregate transaction", pdkLogger.Error(err))
			return err
		}

		eodBalance := beginBalance
		if resp.SumOfDebit+resp.SumOfCredit > 0 {
			eodBalance += resp.SumOfCredit - resp.SumOfDebit
		}

		b.logger.Info(
			ctx, "Balance summary for reference account "+account.ReferenceID.String(),
			pdkLogger.String("source", "CalculateDailyAccountTransaction ("+location.String()+")"),
			pdkLogger.String("type", account.Name), pdkLogger.String("currency", account.Currency),
			pdkLogger.Float64("beginningBalance", beginBalance), pdkLogger.Float64("totalDebit", resp.SumOfDebit),
			pdkLogger.Float64("totalCredit", resp.SumOfCredit), pdkLogger.Float64("endingBalance", eodBalance),
		)

		// Insert daily account transaction
		uuidV7, _ := uuid.NewV7()
		if err := b.dailyAccountTransactionRepo.Upsert(ctx, &dailyAccountTransactionModel.DailyAccountTransaction{
			ID:           uuidV7.String(),
			AccountID:    account.UUID.String(),
			Date:         locationEndAt,
			Timezone:     locationEndAt.Location().String(),
			BegBalance:   beginBalance,
			DebitTrx:     resp.CountOfDebit,
			DebitAmount:  resp.SumOfDebit,
			CreditTrx:    resp.CountOfCredit,
			CreditAmount: resp.SumOfCredit,
			EODBalance:   eodBalance,
		}); err != nil {
			b.logger.Error(ctx, "error when upsert daily account transaction", pdkLogger.Error(err))
			return err
		}
	}

	return nil
}
