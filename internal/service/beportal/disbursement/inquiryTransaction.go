package disbursementService

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (s *DisbursementService) InquiryTransaction(ctx context.Context, request *disbursementModel.InquiryTransaction) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/InquiryTransaction")
	defer segment.End()

	// Get disbursement
	disbursement, err := s.FindByID(ctx, request.DisbursementID)
	if err != nil {
		return nil, err
	}

	// Validate transaction
	if disbursement.MerchantID != request.MerchantID {
		s.logger.Error(ctx, "Merchant not match", pdkLogger.Any("disbursementID", request.DisbursementID))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	if disbursement.TransactionStatus == nil {
		s.logger.Error(ctx, "Transaction not created yet", pdkLogger.Any("disbursementID", request.DisbursementID))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	} else if *disbursement.TransactionStatus != constant.StatusPending {
		s.logger.Error(ctx, "Transaction already in final status", pdkLogger.Any("disbursementID", request.DisbursementID))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrTransactionAlreadyInFinalStatus)
	}
	// End of validation

	err = s.ProcessInquiryTransaction(ctx, disbursement)
	if err != nil {
		return nil, err
	}

	// Find updated disbursement
	updatedDisbursement, errFind := s.FindByID(ctx, request.DisbursementID)
	if errFind != nil {
		return nil, errFind
	}

	return updatedDisbursement, nil
}

// RetryInquirePendingTransactions retries the inquiry of pending transactions within a specified time range.
// This function is used to reduce ops team activity in gathering information about pending transactions and ask  to retry the inquiry process.
// It retrieves pending transactions from the repository, processes each transaction concurrently, and updates the summary of the results.
func (s *DisbursementService) RetryInquirePendingTransactions(ctx context.Context, start, end time.Time) (*disbursementModel.RetryInquireDisbuesementSummary, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/RetryInquirePendingTransactions")
	defer segment.End()

	var (
		eg      = new(errgroup.Group)
		mx      = new(sync.Mutex)
		summary = &disbursementModel.RetryInquireDisbuesementSummary{}
	)

	pendingDisbursements, err := s.disbursementRepo.GetPendingTransactionsBetweenTimeForInquiryTransaction(ctx, start, end)
	if err != nil {
		s.logger.Error(ctx, "failed to get pending transaction between time", pdkLogger.Any("error", err), pdkLogger.String("from", start.String()), pdkLogger.String("to", end.String()))
		return nil, err
	}

	eg.SetLimit(s.totalWorkerPool)

	for _, disbursement := range pendingDisbursements {
		amount, _ := disbursement.Amount.Float64()
		summary.Amount += amount
		summary.Total++

		eg.Go(func() error {
			return s.ProcessPendingTransaction(ctx, disbursement, summary, mx)
		})
	}

	err = eg.Wait()
	if err != nil {
		s.logger.Error(ctx, "error occurred when process pending transaction")
	}

	return summary, nil
}

// ProcessPendingTransaction processes a pending disbursement transaction.
// It updates the summary of succeeded and failed transactions based on the result.
// The function locks the summary using the provided mutex to ensure thread safety.
func (s *DisbursementService) ProcessPendingTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction, summary *disbursementModel.RetryInquireDisbuesementSummary, mx *sync.Mutex) error {
	var (
		err       error
		amount, _ = disbursement.Amount.Float64()
	)

	payoutMutex := s.buildPayoutTransactionMutex(disbursement.UUID)
	if err := payoutMutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "Failed to acquire mutex lock for pending transaction processing", pdkLogger.Error(err))
		return err
	}
	defer func() {
		if _, err := payoutMutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed to release mutex lock for pending transaction processing", pdkLogger.Error(err))
		}
	}()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: disbursement.MerchantID,
		OriginId:    disbursement.UUID,
		From:        "Inquiry-Transaction-Status",
	})

	isFailedTrx := false
	defer func() {
		if isFailedTrx {
			_ = s.self.DecrBeneficiaryPayoutLimit(context.Background(), disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())
		}

		if err != nil {
			mx.Lock()
			summary.TotalFailed++
			summary.AmountFailed += amount
			mx.Unlock()
			return
		}

		mx.Lock()
		summary.TotalSucceeded++
		summary.AmountSucceeded += amount
		mx.Unlock()

	}()

	err = s.ProcessInquiryTransaction(ctx, disbursement)
	if err != nil {
		s.logger.Error(ctx, "error occurred when process pending transaction", pdkLogger.Error(err), pdkLogger.String("disbursementID", disbursement.UUID))
		return err
	}

	updatedDisbursement, err := s.FindByID(ctx, disbursement.UUID)
	if err != nil {
		s.logger.Error(ctx, "error occurred when get updated disbursement", pdkLogger.Error(err), pdkLogger.String("disbursementID", disbursement.UUID))
		return err
	}

	if updatedDisbursement.TransactionStatus == nil {
		s.logger.Info(ctx, "transaction not found", pdkLogger.String("disbursment_id", updatedDisbursement.UUID))
		err = errors.New("transaction is not found")
		return err
	}

	if util.ValueOfPtr(updatedDisbursement.TransactionStatus) == constant.StatusPending {
		s.logger.Info(ctx, "transaction not updated", pdkLogger.String("disbursment_id", updatedDisbursement.UUID))
		err = errors.New("transaction is not updated")
		return err
	}

	if util.ValueOfPtr(updatedDisbursement.TransactionStatus) == constant.StatusFailed {
		isFailedTrx = true
		s.logger.Info(ctx, "transaction failed", pdkLogger.String("disbursment_id", updatedDisbursement.UUID))
		err = errors.New("transaction is failed")
		return err
	}

	return nil
}

// ProcessInquiryTransaction processes an inquiry transaction for a given disbursement.
// It validates the processed transaction, updates the parent status if its a bulk transaction, and sends a callback.
func (s *DisbursementService) ProcessInquiryTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction) error {
	// Inquiry process
	if _, _, err := s.validateProcessedTransaction(ctx, disbursement.UUID, false); err != nil {
		return err
	}

	// Define bulkID
	bulkID := ""
	if disbursement.BulkID != nil {
		bulkID = *disbursement.BulkID
	}

	// Update parent status and send callback
	if err := s.updateParentStatusAndSendCallback(ctx, bulkID, disbursement.UUID); err != nil {
		return err
	}
	return nil
}
