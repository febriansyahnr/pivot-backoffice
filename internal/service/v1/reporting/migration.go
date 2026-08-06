package reportingService

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) MigrateBalanceHistoryToDataReporting(ctx context.Context, startDate, endDate time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reporting/MigrateBalanceHistoryToDataReporting")
	defer segment.End()

	const workerCount = 30

	var (
		wg      = new(sync.WaitGroup)
		chanErr = make(chan error, 1)
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workers, err := ants.NewPoolWithFuncGeneric(workerCount, func(transaction cdcModel.AccountTransaction) {
		defer wg.Done()

		if ctx.Err() != nil {
			return
		}

		now := time.Now().UTC()
		if transaction.RawAdditionalInfo != nil {
			_ = json.Unmarshal([]byte(*transaction.RawAdditionalInfo), &transaction.AdditionalInfo)
		}
		request := reportingModel.UpsertBalanceHistoryRequest{
			Event: &cdcModel.Event[cdcModel.AccountTransaction]{
				Op:    cdcModel.OpCreate,
				After: &transaction,
				TsMs:  now.UnixMilli(),
				TsUs:  now.UnixMicro(),
				TsNs:  now.UnixNano(),
			},
		}
		if err := s.UpsertBalanceHistory(ctx, &request); err != nil {
			chanErr <- err
		}
	})
	if err != nil {
		return fmt.Errorf("unable init worker pool: %w", err)
	}
	defer workers.Release()

	currentDate, nextDate := startDate, startDate
	for currentDate.Before(endDate) {
		nextDate = currentDate.AddDate(0, 0, 1).Add(-time.Second)

		s.logger.Info(ctx, fmt.Sprintf("Starting the process from %s to %s", currentDate.In(tz).Format(time.DateTime), nextDate.In(tz).Format(time.DateTime)))

		transactions, err := s.repo.ListAccountTransactionsForMigration(ctx, currentDate, nextDate)
		if err != nil {
			s.logger.Error(ctx, "Failed to list account transactions for migration", logger.Error(err))
			return err

		} else if len(transactions) == 0 {
			s.logger.Info(ctx, "Transaction not found, data migration process completed")
			return nil
		}

		for _, transaction := range transactions {
			wg.Add(1)
			workers.Invoke(transaction)
		}

		chanWait := make(chan struct{}, 1)
		go func() {
			wg.Wait()
			chanWait <- struct{}{}
		}()

		select {
		case <-chanWait:
			// Proceeding to the next step.
		case err := <-chanErr:
			cancel()
			s.logger.Error(ctx, "Failed to upsert balance history data", logger.Error(err))
			return err
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}
	return nil
}
