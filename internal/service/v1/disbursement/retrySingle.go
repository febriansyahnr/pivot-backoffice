package disbursementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) RetrySingle(ctx context.Context, request *disbursementModel.RetrySingleRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/RetrySingle")
	defer segment.End()

	var err error

	// Get disbursement by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.DisbursementID)
	if err != nil {
		return err
	}
	if disbursement == nil {
		return pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrDisbursementNotFound)
	}

	// Validate merchant
	if disbursement.MerchantID != request.MerchantID {
		err = constant.ErrMerchantIDNotValid
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", disbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrUnauthorized, err)
	}

	// Validate status
	if disbursement.Status != constant.DisbursementStatusApproved {
		err = constant.ErrDisbursementStatusHasNotBeenApproved
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", disbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	// Validate reason, reason cancelled should not proceeded for retry
	if util.ValueOfPtr(disbursement.ReasonType) == constant.DisbursementReasonTypeCancelled {
		err = constant.ErrDisbursementIsCancelled
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", disbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	disbursementIDs := []string{request.DisbursementID}

	// Validate balance
	valid := s.ValidateBalance(ctx, &disbursementModel.ValidateBalanceRequest{
		DisbursementIDs: disbursementIDs,
		MerchantID:      request.MerchantID,
	})
	if !valid {
		err = constant.ErrInsufficientBalance
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", disbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	// Remove INSUFFICIENT_BALANCE reason
	err = s.disbursementRepo.UpdateReasonByIDs(ctx, disbursementIDs, "", "")
	if err != nil {
		return err
	}

	transactionSummary, err := s.disbursementRepo.GetActionTransactionSummary(ctx, request.MerchantID, disbursementIDs)
	if err != nil {
		s.logger.Error(ctx, "failed to get action transaction summary", logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	if transactionSummary == nil || transactionSummary.Total != len(disbursementIDs) {
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrMerchantNotAllowedPerformAction)
	}

	_, err = s.self.ValidateDailyTransactionLimit(ctx, request.MerchantID, transactionSummary.TotalAmount)
	if err != nil {
		s.logger.Error(ctx, "failed to validate daily transaction limit", logger.Error(err))
		return err
	}

	// Restore parent merchant context so the fee on behalf via parent logic
	if disbursement.MetadataObj.OnBehalf != nil && disbursement.MetadataObj.OnBehalf.ParentMerchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, disbursement.MetadataObj.OnBehalf.ParentMerchantId)
	}

	transactionId, transactionFeeId := "", ""
	if transactionId, transactionFeeId, err = s.self.CreatePendingOrchestratorTransaction(ctx, disbursement); err != nil {
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	defer func() {
		if err != nil {
			errReset := s.self.DeleteDailyTransactionLimit(ctx, request.MerchantID)
			if errReset != nil {
				s.logger.Error(ctx, "error when reset daily limit", logger.Error(errReset))
			}
		}
	}()

	// Validate beneficiary payout limit
	err = s.self.ValidateBankAccountAndUpdateTransaction(ctx, disbursement, &orchestraModel.TransactionAndFeeObject{
		TransactionID: transactionId,
		FeeID:         transactionFeeId,
		MerchantID:    disbursement.MerchantID,
	})
	if err != nil {
		return err
	}

	s.triggerPublishBatchProcess(ctx, "", disbursementIDs, nil)

	return nil
}
