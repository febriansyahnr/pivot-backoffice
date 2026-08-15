package disbursementService

import (
	"context"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	routingProcessorModelPriority "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/processorPriority"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) getTransactionByExternalID(ctx context.Context, externalID string) (*orchestratorModel.TransactionAndFeeObject, *disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/getTransactionByExternalID")
	defer segment.End()

	var existedTransactionIDInStr, existedFeeIDInStr string

	existedTransaction, err := s.accountTransactionRepo.FindByID(ctx, externalID)
	if err != nil {
		return nil, nil, err
	}
	if existedTransaction == nil {
		return nil, nil, constant.ErrDataNotFound
	}

	existedTransactionIDInStr = existedTransaction.UUID.String()
	disbursementID := existedTransaction.ReferenceID

	// Find disbursement by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, disbursementID)
	if err != nil {
		return nil, nil, err

	} else if disbursement == nil {
		return nil, nil, constant.ErrDisbursementNotFound
	}

	// Find fee from existed transaction
	existedFee, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeFee)
	if err != nil {
		return nil, nil, err
	}
	if existedFee == nil {
		return nil, nil, constant.ErrDataNotFound
	}
	existedFeeIDInStr = existedFee.UUID.String()

	orchestratorTransaction := &orchestratorModel.TransactionAndFeeObject{
		MerchantID:    existedTransaction.MerchantID.String(),
		TransactionID: existedTransactionIDInStr,
		FeeID:         existedFeeIDInStr,
	}
	if additionalInfo, ok := existedFee.AdditionalInfoObj.(orchestratorModel.FeeTransactionMetadataObject); ok {
		orchestratorTransaction.TransferFeeID = additionalInfo.TransferId
	}
	return orchestratorTransaction, disbursement, nil
}

// ProcessUpdateTransferStatus process update transfer status
func (s *DisbursementService) ProcessUpdateTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ProcessUpdateTransferStatus")
	defer segment.End()

	var reasonType, reasonDesc string

	orchestratorTransaction, disbursement, err := s.getTransactionByExternalID(ctx, request.ExternalID)
	if err != nil {
		s.logger.Error(ctx, "[ProcessUpdateTransferStatus] Cannot get transaction by external id", logger.Error(err), logger.String("externalID", request.ExternalID))
		return err
	}

	isDana := strings.Compare(request.ProcessorReference, constant.DanaPGProcessor) == 0
	// Request callback from Dana doesn't contain Amount
	// bypass validation amount for Dana
	if !isDana {
		// validation amount
		requestAmount := request.Amount.ToDecimal()
		disbursementAmount := disbursement.Amount

		if requestAmount.Cmp(disbursementAmount) != 0 {
			s.logger.Error(ctx, "[ProcessUpdateTransferStatus] Amount mismatch",
				logger.String("externalID", request.ExternalID))
			return constant.ErrInvalidRequestPayload
		}
	}

	accountTransactionStatus := translateSnapcoreStatus(request.Status)
	if request.Status != constant.StatusSuccess {
		reasonType = constant.ReasonTypeOtherReason
		reasonDesc = request.ResponseMessage
	}

	reasonType, reasonDesc = s.handlingReasonTypeAndDescFromTransfer(request.ResponseCode, request.ResponseMessage)
	if reasonType == constant.ReasonTypeInsufficientEscrowFund {
		accountTransactionStatus = constant.StatusPending
	}

	if request.Status == "" || slices.Contains([]string{"01", "02", "03", "07"}, request.LatestTransactionStatus) {
		accountTransactionStatus = constant.StatusPending
		reasonType = constant.ReasonTypePayoutDelayed
		reasonDesc = constant.ReasonDescPayoutDelayed
	}

	if rt, ok := request.AdditionalInfo["reasonType"].(string); ok && rt == constant.ReasonTypePayoutCutOffTime {
		accountTransactionStatus = constant.StatusPending
		reasonType = constant.ReasonTypePayoutCutOffTime
		reasonDesc = constant.ReasonDescPayoutCutOffTime
	}

	// Begin Tx for update disbursement status
	ctxTrx, trxErr := s.disbursementRepo.BeginTransaction(ctx)
	if trxErr != nil {
		return trxErr
	}

	txCompleted, isFailedTrx := false, false
	defer func() {
		if txCompleted && isFailedTrx {
			_ = s.self.DecrDailyTransactionLimit(context.Background(), disbursement.MerchantID, disbursement.Amount.InexactFloat64())
			_ = s.self.DecrBeneficiaryPayoutLimit(context.Background(), disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())
		}
	}()

	defer func() {
		if !txCompleted {
			if trxErr = s.disbursementRepo.RollbackTransaction(ctxTrx); trxErr != nil {
				return
			}
		}
	}()

	if accountTransactionStatus == constant.StatusSuccess {
		// Update disbursement set processorReferenceId = snapCoreResp.UUID
		if errUpdate := s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(
			ctxTrx, disbursement.UUID, request.UUID, request.BankReferenceNo); errUpdate != nil {
			return errUpdate
		}

		if errUpdateProcRef := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx,
			orchestratorTransaction.TransactionID, request.ProcessorReference, request.UUID, request.GetReconReferenceNo()); errUpdateProcRef != nil {
			return errUpdateProcRef
		}

		// Update transaction and fee status
		if disbursement.MetadataObj.FeeDetail.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
			orchestratorTransaction.FeeID = ""
		}
		reasonType = ""
		reasonDesc = ""

		// Increment deferred LADDER tiering counter now that the payout has succeeded
		feeDetail := disbursement.MetadataObj.FeeDetail
		if feeDetail.LadderCounterKey != "" {
			s.feeSvc.IncrementLadderCounter(ctx, feeDetail.LadderCounterKey, feeDetail.LadderCounterIncrement)
		}
	}

	if errUpdate := s.updateTransactionStatusWithHistory(ctxTrx, orchestratorTransaction,
		accountTransactionStatus, &reasonType, &reasonDesc, disbursement.UUID); errUpdate != nil {
		return errUpdate
	}

	// Commit Tx
	if trxErr = s.disbursementRepo.CommitTransaction(ctxTrx); trxErr != nil {
		return trxErr
	}
	txCompleted = true

	// Suppress callbacks for cutoff-held payouts — they will resume automatically
	if reasonType == constant.ReasonTypePayoutCutOffTime {
		return nil
	}

	// Update bulk disbursement status and send callback when conditions are met
	if disbursement.BulkID != nil {
		// Send callback for delayed payout
		if accountTransactionStatus == constant.StatusPending && reasonType == constant.ReasonTypePayoutDelayed {
			return s.sendCallback(ctx, *disbursement.BulkID, disbursement.MerchantID, constant.BulkDisbursementStatusPending, constant.CallbackEventPayoutDelayed)
		}

		// Trigger check count in progress, if count = 0 then update bulkDisbursement status to DONE
		// Use context.Background() because the existing context was used for create transaction :)
		if errUpdate := s.updateParentStatusAndSendCallback(ctx, *disbursement.BulkID, disbursement.UUID); errUpdate != nil {
			return errUpdate
		}
	}

	// if status is FAILED, the system will check for next priority, and do execute for a new transfer
	// if next priority is found, the system will execute the next transfer process
	if request.Status == constant.StatusFailed {
		// check for next priority is exists?
		processor := s.getProcessorNextPriority(ctx, request.ProcessorReference, disbursement.MerchantID)
		if processor != nil {
			ctx = context.WithValue(ctx, constant.CtxProcessorName, processor.ProcessorName)
			err = s.ProcessBankTransferAndUpdateTransaction(ctx, disbursement, orchestratorTransaction)
		}
	}

	isFailedTrx = accountTransactionStatus == constant.StatusFailed

	return err
}

func translateSnapcoreStatus(status string) string {
	switch status {
	case constant.SnapCoreBankTransferStatusSuccess:
		return constant.StatusSuccess
	case constant.SnapCoreBankTransferStatusFailed:
		return constant.StatusFailed
	case constant.SnapCoreBankTransferStatusPending:
		return constant.StatusPending
	default:
		return constant.StatusFailed
	}
}

func (s *DisbursementService) getProcessorNextPriority(ctx context.Context,
	processorReferenceName string,
	merchantID string,
) *routingProcessorModelPriority.ProcessorPriority {
	processorList := s.routingProcessorSvc.GetProcessorList(ctx, merchantID)
	var processor routingProcessorModelPriority.ProcessorPriority
	for i, routeConfig := range processorList {
		if routeConfig.ProcessorName == processorReferenceName {
			if i+1 < len(processorList) && processorList[i+1].IsActive {
				processor = processorList[i+1]
				return &processor
			}
		}
	}

	return nil
}

func (s *DisbursementService) handlingReasonTypeAndDescFromTransfer(
	responseCode, responseMessage string,
) (reasonType, reasonDesc string) {
	switch {
	case util.IsPatternMatch(constant.SnapCoreResponseCodeInsufficientFundPattern, responseCode):
		reasonType = constant.ReasonTypeInsufficientEscrowFund
		reasonDesc = responseMessage
	case util.IsPatternMatch(constant.SnapCoreResponseCodeInactiveAccountPattern, responseCode):
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseInactiveAccountMessage
	case util.IsPatternMatch(constant.SnapCoreResponseCodeDormantAccountPattern, responseCode):
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseDormantAccountMessage
	case util.IsPatternMatch(constant.SnapCoreResponseCodeInvalidAccountPattern, responseCode):
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseInvalidAccountMessage
	}

	return
}
