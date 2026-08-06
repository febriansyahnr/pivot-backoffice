package paymentService

import (
	"context"
	"fmt"
	"time"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) ProcessInvestigationMonthlyReconciliation(ctx context.Context, request model.MonthlyReconciliationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ProcessInvestigationMonthlyReconciliation")
	defer segment.End()

	transactions, err := s.paymentRepo.CalculateInvestigationMonthlyReconciliation(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "Failed to calculate investigation monthly reconciliation", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	} else if len(transactions) == 0 {
		s.logger.Info(ctx, "There are no transactions resulting from the investigation that need to be reconciled")
		return nil
	}

	successCount, failedCount := 0, 0
	for _, transaction := range transactions {
		if err := s.processReconciliationFromInvestigation(ctx, transaction); err == nil {
			successCount++
		} else {
			failedCount++
			s.logger.Error(ctx, "Failed to process reconciliation from payment investigation", logger.Error(err), logger.Any("transaction", transaction))
		}
	}

	s.logger.Info(ctx, "Process monthly reconciliation completed", logger.Any("summaries", map[string]any{
		"total":   len(transactions),
		"success": successCount,
		"failed":  failedCount,
	}))
	return nil
}

func (s *PaymentService) processReconciliationFromInvestigation(ctx context.Context, transaction model.CalculateInvestigationMonthlyReconciliation) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/processReconciliationFromInvestigation")
	defer segment.End()

	ctxTrx, err := s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	defer func() {
		if err == nil {
			return
		}
		if er := s.paymentRepo.RollbackTransaction(ctxTrx); er != nil {
			s.logger.Error(ctx, "Failed to rollback transaction", logger.Error(er))
		}
	}()

	id := util.GenerateUUID()
	date := time.Now().UTC()

	request := transaction.ToPaymentInvestigationMonthlyReconciliation(id.String(), date)

	if err = s.paymentRepo.InsertInvestigationMonthlyReconciliation(ctxTrx, request); err != nil {
		return fmt.Errorf("create payment investigation monthly reconciliation: %s", err)
	}

	if err = s.paymentRepo.UpdatePaymentInvestigationReconciliation(ctxTrx, request); err != nil {
		return fmt.Errorf("update payment investigation reconciliation: %s", err)
	}

	if err = s.orchestratorSvc.CreateAccountTransaction(ctxTrx, transaction.ToCreateAccountTransactionRequest(request.UUID, request.Date)); err != nil {
		return fmt.Errorf("create account transaction: %s", err)
	}

	return s.paymentRepo.CommitTransaction(ctxTrx)
}
