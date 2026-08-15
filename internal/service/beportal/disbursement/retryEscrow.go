package disbursementService

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) RetryDueToInsufficientEscrowFund(ctx context.Context, request *disbursementModel.RetryTransaction) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/RetryDueToInsufficientEscrowFund")
	defer segment.End()

	// Get disbursement
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.DisbursementID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if disbursement == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound)
	}

	// Validate transaction
	if disbursement.MerchantID != request.MerchantID {
		s.logger.Error(ctx, "Merchant not match", logger.String("disbursementID", request.DisbursementID))
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	if disbursement.TransactionStatus == nil {
		s.logger.Error(ctx, "Transaction not created yet", logger.String("disbursementID", request.DisbursementID))
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}
	// End of validation

	payoutMutex := s.buildPayoutTransactionMutex(disbursement.UUID)
	if err := payoutMutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "Failed to acquire mutex lock for retry transaction processing", logger.Error(err))
		return err
	}
	defer func() {
		if _, err := payoutMutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed to release mutex lock for retry transaction processing", logger.Error(err))
		}
	}()

	// Log forceFailed parameter (passthrough to SNAP Core)
	s.logger.Info(ctx, "Retry transaction with forceFailed parameter",
		logger.String("disbursementID", request.DisbursementID),
		logger.Bool("forceFailed", request.ForceFailed))

	// Store forceFailed in context for downstream passthrough
	ctx = context.WithValue(ctx, constant.CtxForceFailed, request.ForceFailed)
	ctx = context.WithValue(ctx, constant.CtxFromRetry, true)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: request.MerchantID,
		OriginId:    request.DisbursementID,
		From:        "Retry-Payout-Transaction",
	})

	// Restore parent merchant context so the fee on behalf via parent logic
	if disbursement.MetadataObj.OnBehalf != nil && disbursement.MetadataObj.OnBehalf.ParentMerchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, disbursement.MetadataObj.OnBehalf.ParentMerchantId)
	}

	// Define bulkID
	bulkID := ""
	if disbursement.BulkID != nil {
		bulkID = *disbursement.BulkID
	}

	accountTransaction, err := s.accountTransactionRepo.FindByReference(ctx, disbursement.UUID, constant.ReferenceDisbursement)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}
	if accountTransaction == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	reason := ""
	isAllowedToRetry := false
	if !request.BypassProcessorCheck {
		checkRetryResult, errCheckRetry := s.snapCoreRepo.CheckAllowedToRetry(ctx, snapCoreModel.CheckAllowedToRetryRequest{
			ExternalID: accountTransaction.UUID.String(),
			MerchantId: request.MerchantID,
			Force:      request.ForceRetry,
		})
		if errCheckRetry != nil {
			return errCheckRetry
		}
		isAllowedToRetry = checkRetryResult.Allowed
		reason = checkRetryResult.Reason
	} else {
		isAllowedToRetry = true
	}

	if !isAllowedToRetry {
		return pkgErrors.New(response.HttpErrRequest, errors.New(reason))
	}
	// Process retry without create new account transaction.
	if errProcess := s.Process(ctx, request.DisbursementID, true); errProcess != nil {

		// Update parent status and send callback (FAILED)
		if errUpdate := s.updateParentStatusAndSendCallback(ctx, bulkID, disbursement.UUID); errUpdate != nil {
			return errUpdate
		}

		return errProcess
	}

	// Update parent status and send callback (SUCCESS)
	if errUpdate := s.updateParentStatusAndSendCallback(ctx, bulkID, disbursement.UUID); errUpdate != nil {
		return errUpdate
	}

	return nil
}
