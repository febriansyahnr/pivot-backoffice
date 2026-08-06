package disbursementService

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

// ChangeDisbursementTransactionStatus processes the change of status for multiple disbursement transactions concurrently.
// It takes a context and a request containing the disbursement IDs and other necessary information.
// The function returns a slice of responses indicating the result of each transaction status change.
// Will return []disbursementModel.ChangeDisbursementTransactionStatusResponse: A slice of responses indicating the result of each transaction status change.
func (s *DisbursementService) ChangeDisbursementTransactionStatus(ctx context.Context, req disbursementModel.ChangeDisbursementTransactionStatusRequest) []disbursementModel.ChangeDisbursementTransactionStatusResponse {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ChangeDisbursementTransactionStatus")
	defer segment.End()

	var (
		eg     errgroup.Group
		mx     sync.Mutex
		result = []disbursementModel.ChangeDisbursementTransactionStatusResponse{}
	)

	eg.SetLimit(s.totalWorkerPool)
	for _, disbursementID := range req.DisbursementIDS {
		eg.Go(func() error {
			isUpdated := true
			reason := "ok"

			err := s.ProcessChangeDisbursementTansactionStatus(ctx, disbursementID, req)
			if err != nil {
				isUpdated = false
				reason = err.Error()
				s.logger.Error(ctx, "failed to process change disbursement transaction status", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
			}

			mx.Lock()
			result = append(result, disbursementModel.ChangeDisbursementTransactionStatusResponse{
				DisbursementID: disbursementID,
				Updated:        isUpdated,
				Reason:         reason,
			})
			mx.Unlock()

			return err
		})
	}

	_ = eg.Wait()

	return result
}

// ProcessChangeDisbursementTansactionStatus processes the change of disbursement transaction status.
// It retrieves the disbursement and ledger information, updates the transaction status, and handles
// the transaction commit and rollback and will process the request using 3 go routine.
// error: An error object if any error occurs during the process, otherwise nil.
func (s *DisbursementService) ProcessChangeDisbursementTansactionStatus(ctx context.Context, disbursementID string, req disbursementModel.ChangeDisbursementTransactionStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ProcessChangeDisbursementTansactionStatus")
	defer segment.End()

	orchestratorTransaction := &orchestratorModel.TransactionAndFeeObject{}

	disbursement, err := s.disbursementRepo.FindByID(ctx, disbursementID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find disbursement", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
		return err
	}

	if disbursement == nil {
		s.logger.Error(ctx, "Disbursement not found", pdkLog.String("disbursementID", disbursementID))
		return constant.ErrDataNotFound
	}

	ledger, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "Failed to find ledger", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
		return err
	}

	if ledger == nil {
		s.logger.Error(ctx, "Ledger not found", pdkLog.String("disbursementID", disbursementID))
		return constant.ErrDataNotFound
	}

	if ledger.Status == constant.StatusSuccess {
		return constant.ErrTransactionAlreadySucceeded
	}

	orchestratorTransaction.MerchantID = ledger.MerchantID.String()
	orchestratorTransaction.TransactionID = ledger.UUID.String()

	feeLedger, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeFee)
	if err != nil {
		s.logger.Error(ctx, "Failed to find fee ledger", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
		return err
	}

	var feeMetadata feeModel.FeeMetadataObject
	if feeLedger != nil {
		orchestratorTransaction.FeeID = feeLedger.UUID.String()

		_ = json.Unmarshal(feeLedger.AdditionalInfo.JSONText, &feeMetadata)

		orchestratorTransaction.TransferFeeID = feeMetadata.TransferId
	}

	// Increment deferred LADDER tiering counter when status is changed to SUCCESS
	if req.Status == constant.StatusSuccess && feeMetadata.LadderCounterKey != "" {
		s.feeSvc.IncrementLadderCounter(ctx, feeMetadata.LadderCounterKey, feeMetadata.LadderCounterIncrement)
	}

	// Begin Tx
	trxCtx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	isCompleted := false
	defer func() {
		// when the transaction is completed, we need to remove the daily transaction limit so it can recalcalculated again
		if !isCompleted {
			return
		}

		err := s.redisExt.Del(
			context.Background(),
			fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, ledger.MerchantID, constant.DisbursementDailyLimitMerchant),
			fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, ledger.MerchantID, constant.DisbursementDailyLimitMerchantPlatform)).Err()
		if err != nil {
			s.logger.Error(ctx, "Failed to delete daily transaction limit", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
		}
	}()

	defer func() {
		if !isCompleted {
			err = s.disbursementRepo.RollbackTransaction(trxCtx)
			if err != nil {
				s.logger.Error(ctx, "Failed to rollback transaction", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
			}
		}
	}()

	err = s.updateTransactionStatusWithHistory(trxCtx, orchestratorTransaction, req.Status, req.ReasonType, req.ReasonDescription, disbursementID)
	if err != nil {
		return err
	}

	// When a reference number is provided, persist it as the bank reference number
	// for this disbursement. Runs inside the same transaction so it rolls back together
	// with the status update on failure.
	if strings.TrimSpace(req.ReferenceNumber) != "" {
		if err := s.disbursementRepo.UpdateBankReferenceNo(trxCtx, disbursementID, req.ReferenceNumber); err != nil {
			return err
		}
	}

	if ledger.ProcessorReference == constant.SnapCoreProcessor {
		err := s.snapCoreRepo.UpdateBankTransferStatus(trxCtx, snapCoreModel.UpdateBankTransferStatusRequest{
			ExternalID: ledger.UUID.String(),
			Status:     req.Status,
		})

		if err != nil {
			return err
		}
	}

	// Commit Tx
	err = s.disbursementRepo.CommitTransaction(trxCtx)
	if err != nil {
		return err
	}

	err = s.updateParentStatusAndSendCallback(ctx, util.ValueOfPtr(disbursement.BulkID), disbursementID)
	if err != nil {
		s.logger.Error(ctx, "Failed to update parent status and send callback", pdkLog.String("disbursementID", disbursementID), pdkLog.Error(err))
	}

	s.logger.Info(ctx, "update disbursement transaction status success", pdkLog.String("disbursementID", disbursementID))
	isCompleted = true
	return nil
}
