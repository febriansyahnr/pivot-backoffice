package disbursementService

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) RetryBulk(ctx context.Context, request *disbursementModel.RetryBulkRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/RetryBulk")
	defer segment.End()

	// Get bulk_disbursement by ID
	bulkDisbursement, err := s.disbursementRepo.FindBulkDisbursementByID(ctx, request.BulkDisbursementID)
	if err != nil {
		return err
	}
	if bulkDisbursement == nil {
		return pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// Validate merchant
	if bulkDisbursement.MerchantID != request.MerchantID {
		err = constant.ErrMerchantIDNotValid
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", bulkDisbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrUnauthorized, err)
	}

	// Validate status
	if bulkDisbursement.Status != constant.StatusPending {
		err = constant.ErrDisbursementStatusHasNotBeenApproved
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", bulkDisbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrUnprocessableContent, err)
	}

	var disbursementIDs []string

	// Get child disbursement by bulkDisbursementId
	disbursements, err := s.disbursementRepo.GetAllDisbursementByBulkID(ctx, request.BulkDisbursementID)
	if err != nil {
		return err
	}

	for _, disbursement := range disbursements {
		reasonType := ""
		if disbursement.ReasonType != nil {
			reasonType = *disbursement.ReasonType
		}

		err = json.Unmarshal(disbursement.Metadata.JSONText, &disbursement.MetadataObj)
		if err != nil {
			s.logger.Error(ctx, "failed to unmarshal disbursement metadata", logger.Error(err), logger.String("id", disbursement.UUID))
		}

		if disbursement.Status == constant.DisbursementStatusApproved && reasonType == constant.DisbursementReasonTypeInsufficientBalance {
			disbursementIDs = append(disbursementIDs, disbursement.UUID)
		}
	}

	// Validate balance
	valid := s.ValidateBalance(ctx, &disbursementModel.ValidateBalanceRequest{
		DisbursementIDs: disbursementIDs,
		MerchantID:      request.MerchantID,
	})
	if !valid {
		err = constant.ErrInsufficientBalance
		s.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("id", bulkDisbursement.UUID))
		return pkgErrors.New(httpResponse.HttpErrForbidden, err)
	}

	// Remove INSUFFICIENT_BALANCE reason
	err = s.disbursementRepo.UpdateReasonByIDs(ctx, disbursementIDs, "", "")
	if err != nil {
		return err
	}

	// Update bulk disbursement status to IN_PROGRESS
	if err = s.disbursementRepo.UpdateBulkDisbursementStatusByID(ctx, bulkDisbursement.UUID, constant.BulkDisbursementStatusInProgress); err != nil {
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

	approvalResErr := disbursementModel.ApprovalResultErr{}
	for _, disbursement := range disbursements {
		var (
			err                             error
			transactionId, transactionFeeId = "", ""
		)

		// Restore parent merchant context per disbursement so the fee on behalf
		iterCtx := ctx
		if disbursement.MetadataObj.OnBehalf != nil && disbursement.MetadataObj.OnBehalf.ParentMerchantId != "" {
			iterCtx = context.WithValue(iterCtx, constant.CtxParentMerchantId, disbursement.MetadataObj.OnBehalf.ParentMerchantId)
		}

		if transactionId, transactionFeeId, err = s.self.CreatePendingOrchestratorTransaction(iterCtx, disbursement); err != nil {
			s.logger.Error(ctx, "error when create pending orchestrator transaction", logger.Any("details", request))
		}

		// Validate beneficiary payout limit
		if errBeneficiaryLimit := s.self.ValidateBankAccountAndUpdateTransaction(iterCtx, disbursement, &orchestraModel.TransactionAndFeeObject{
			TransactionID: transactionId,
			FeeID:         transactionFeeId,
			MerchantID:    request.MerchantID,
		}); errBeneficiaryLimit != nil {
			approvalResErr.BeneficiaryLimitExceeded = append(approvalResErr.BeneficiaryLimitExceeded, disbursementModel.ApprovalValidation{
				Amount:    disbursement.Amount.InexactFloat64(),
				AccountNo: disbursement.BeneficiaryAccountNo,
				Error:     errBeneficiaryLimit,
			})
			_ = s.self.DecrDailyTransactionLimit(ctx, request.MerchantID, disbursement.Amount.InexactFloat64())
		}
	}

	// Trigger process disbursement in async (bulk)
	s.triggerPublishBatchProcess(ctx, bulkDisbursement.UUID, disbursementIDs, nil)

	if len(approvalResErr.BeneficiaryLimitExceeded) == len(disbursementIDs) {
		return pkgErrors.New(httpResponse.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions)
	}

	if len(approvalResErr.BeneficiaryLimitExceeded) > 0 {
		return &approvalResErr
	}

	return nil
}
