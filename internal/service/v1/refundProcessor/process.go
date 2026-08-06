package refundProcessorService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *RefundProcessor) Process(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/Process")
	defer span.End()

	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)
	if parentMerchantID == "" {
		merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "error when find merchant", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		if merchant.ParentID.Valid && merchant.ParentID.String != "" {
			ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
		}
	}

	// Validate refund availability by refund status
	if request.Status != constant.RefundStatusPending {
		s.logger.Warn(ctx, "[RefundProcess] Refund not in pending status")
		return pkgErrs.New(response.HttpStatusErrorUnprocessableContent, constant.ErrRefundNotInPendingStatus)
	}

	// Implement exclusive lock for process refund by refundID
	queueKey := fmt.Sprintf("backend-portal:refund-process:%s", request.RefundID)
	if ok, errLock := s.redis.SetNX(ctx, queueKey, true, 5*time.Minute).Result(); errLock != nil {
		s.logger.Error(ctx, "[RefundProcess] Set exclusive queue with key "+queueKey, logger.Error(errLock))
		return pkgErrs.New(response.HttpErrDatabase, errLock)

	} else if !ok {
		return pkgErrs.New(response.HttpErrDupCheck, constant.ErrRefundIsBeingProcessed)
	}

	ctxTx, err := s.refundRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "[RefundProcess] Begin transaction", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	isCompleted := false
	defer func() {
		if isCompleted {
			return
		}
		if e := s.refundRepo.RollbackTransaction(ctxTx); e != nil {
			s.logger.Error(ctx, "[RefundProcess] Rollback session transaction", logger.Error(e))
		}
	}()

	// Func update refund status to WAITING_BANK_TRANSFER
	updateRefundToWaitingBankTransfer := func() {
		request.Refund.Status = constant.RefundStatusWaitingBankTransfer
		request.Refund.DestinationType = constant.RefundDestinationTypeAccount
		errUpdate := s.refundRepo.UpdateData(ctx, request.Refund)
		if errUpdate != nil {
			s.logger.Error(ctx, "[RefundProcess] Failed to update refund status to WAITING_BANK_TRANSFER", logger.Error(errUpdate))
		}

		errCallback := s.refundSvc.SendCallback(ctx, request.RefundID, request.MerchantID)
		if errCallback != nil {
			s.logger.Error(ctx, "[RefundProcess] Send callback", logger.Error(errCallback))
		}
	}

	// Process through channel or transfer
	// TODO: reduce code complexity and issues
	err, destinationType := func() (err error, destinationType string) {
		destinationType = constant.RefundDestinationTypeChannel

		charge, errdb := s.orchestratorSvc.FindByID(ctx, request.PaymentChargeID)
		if errdb != nil {
			s.logger.Error(ctx, "[RefundProcess] Failed to find payment charge", logger.Error(errdb))
			err = pkgErrs.New(response.HttpErrUnprocessableContent, errdb)
			return
		}
		if charge.SettlementStatus.String == constant.SettlementStatusPending {
			s.logger.Info(ctx, "[RefundProcess] Payment charge is unsettled, skip check balance")
			ctxTx = context.WithValue(ctxTx, constant.CtxSetBypassBalanceCheckTransaction, true)
		}

		switch request.Method {
		case constant.RefundMethodTransferOnly:
			updateRefundToWaitingBankTransfer()
			return s.bankTransfer.Process(ctxTx, request), constant.RefundDestinationTypeAccount
		case constant.RefundMethodAuto:
			switch request.PaymentMethodType {
			case constant.UnifiedPaymentMethodCard, paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
				err = s.card.Process(ctxTx, request)
			case constant.UnifiedPaymentMethodQris, paymentConstant.PAYMENT_METHOD_QRIS:
				err = s.qris.Process(ctxTx, request)
			case constant.UnifiedPaymentMethodEWallet:
				err = s.ewallet.Process(ctxTx, request)
			case constant.UnifiedPaymentMethodVA:
				updateRefundToWaitingBankTransfer()
				return s.bankTransfer.Process(ctxTx, request), constant.RefundDestinationTypeAccount
			default:
				return fmt.Errorf("[RefundProcess] unsupported payment method type for CHANNEL refund method"), destinationType
			}

			if err != nil {
				if constant.IsDirectPSP(request.PaymentMethodChannelType) {
					return err, destinationType
				}

				if request.Refund != nil && request.Refund.MetadataObj.TransferDestination == nil {
					s.logger.Info(ctx, "[RefundProcess] Refund does not have transfer destination", logger.String("refundId", request.RefundID))
					return err, destinationType
				}

				updateRefundToWaitingBankTransfer()
				return s.bankTransfer.Process(ctxTx, request), constant.RefundDestinationTypeAccount
			}
			return nil, destinationType
		default:
			return fmt.Errorf("[RefundProcess] unsupported refund method"), destinationType
		}
	}()

	if request.DestinationType != destinationType {
		request.DestinationType = destinationType
	}

	if err != nil {
		if errors.Is(err, constant.ErrBankTransferStillInPending) {
			s.logger.Info(ctx, "[RefundProcess] Bank transfer process still pending")
			return nil
		}

		// Update status refund to failed
		request.Status = constant.RefundStatusFailed
		if errUpdate := s.refundRepo.UpdateData(ctxTx, request.Refund); errUpdate != nil {
			return pkgErrs.New(response.HttpErrDatabase, errUpdate)
		}

		// Record refund failed status
		s.refundSvc.RecordRefundStatusHistory(ctx, request.RefundID, constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed)

		// Update status refund ledger to failed
		reasonType, reasonDesc := constant.ReasonTypeOtherReason, ""
		if request.RefundLedgerReasonType != nil {
			reasonType = *request.RefundLedgerReasonType
		}
		if request.RefundLedgerReasonDesc != nil {
			reasonDesc = *request.RefundLedgerReasonDesc
		}
		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransactionByReferenceID(ctxTx, request.RefundLedgerReferenceID, constant.StatusFailed, &reasonType, &reasonDesc); errUpdate != nil {
			s.logger.Error(ctx, "[RefundProcess] Update status account transaction (failed)", logger.Error(errUpdate))
		}

		if err = s.refundRepo.CommitTransaction(ctxTx); err != nil {
			s.logger.Error(ctx, "[RefundProcess] Commit session transaction", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		isCompleted = true

		errCallback := s.refundSvc.SendCallback(ctx, request.RefundID, request.MerchantID)
		if errCallback != nil {
			s.logger.Error(ctx, "[RefundProcess] Send callback when failed ro process", logger.Error(errCallback))
		}

		return err
	}

	// Update refund to success and handle associated operations
	err = s.updateRefundToSuccess(ctxTx, ctx, request)
	if err != nil {
		return err
	}

	if err = s.refundRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "Commit session transaction", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	// Force settlement of unsettled payment
	if request.PaymentChargeSettlementStatus == constant.StatusPending {
		if errSettle := s.settlementSvc.ProcessSettlement(ctx, &settlementModel.ProcessSettlementRequest{
			MerchantID:           request.MerchantID,
			Type:                 constant.SettlementTransaction,
			TransactionID:        request.PaymentChargeID,
			FeeTransactionID:     request.PaymentFeeID,
			ByPassSettlementHold: true,
		}); errSettle != nil {
			s.logger.Error(ctx, "[RefundProcess] failed to force settlement", logger.Error(err))
		}
	}

	errCallback := s.refundSvc.SendCallback(ctx, request.RefundID, request.MerchantID)
	if errCallback != nil {
		s.logger.Error(ctx, "[RefundProcess] Send callback on success", logger.Error(errCallback))
	}

	return err
}

// ChargeSubMerchantToMerchant processes a refund by charging a sub-merchant and transferring funds to the parent merchant.
// It performs the following steps:
// Gets the transaction fee information for the refund, Creates a transfer request from the sub-merchant to the parent merchant
// Then Executes the transfer, Updates the transfer status to success
func (s *RefundProcessor) ChargeSubMerchantToMerchant(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)

	if constant.IsDirectPSP(request.PaymentMethodChannelType) {
		s.logger.Info(ctx, "[RefundProcess] Payment method channel type is facilitator", logger.String("refundId", request.RefundID))
		return nil
	}

	feeMetadata, err := s.feeSvc.GetTransactionFeeOnBehalf(ctx, &feeModel.GetTrxFeeOnBehalfRequest{
		MerchantId:        parentMerchantID,
		SubMerchantId:     request.MerchantID,
		Reference:         constant.ReferenceRefund,
		TransactionAmount: request.Amount,
	})
	if err != nil {
		s.logger.Error(ctx, "[RefundProcess] failed to get fee on behalf", logger.Error(err))
		return err
	}

	if feeMetadata.FinalAmount <= 0 {
		s.logger.Warn(ctx, "[RefundProcess] merchant fee on behalf less than 0", logger.String("merchant_id", request.MerchantID))
		return nil
	}

	transferReq := &transfer.TransferRequest{
		SourceMerchantID: util.ParseUUID(request.MerchantID),
		RecipientID:      parentMerchantID,
		ParentMerchantID: util.ParseUUID(parentMerchantID),
		ReferenceID:      request.UUID,
		TransferType:     constant.MoneyFlowDirect,
		Amount:           feeMetadata.FinalAmount,
		Remarks:          fmt.Sprintf("Refund fee for ref_id: %s", request.ClientReferenceID),
		Usecase:          constant.TypeRefund,
	}

	transfer, err := s.transferSvc.Transfer(ctx, transferReq)
	if err != nil {
		s.logger.Error(ctx, "[RefundProcess] Failed transfer to the parent merchant", logger.Error(err))
		return err
	}

	// Update transfer status to success
	if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, transfer.UUID.String(), constant.StatusSuccess, nil, nil); errUpdate != nil {
		s.logger.Error(ctx, "[RefundProcess] Update status account transaction (success)", logger.Error(errUpdate))
		return pkgErrs.New(response.HttpErrDatabase, errUpdate)
	}

	s.logger.Info(ctx, "[RefundProcess] Success transfer to the parent merchant", logger.String("transfer_id", transfer.UUID.String()))
	return nil
}

// CreditMDRToMerchant calculates and processes the refund of the payment fee (Merchant Discount Rate) back to the merchant,
// determining the recipient (parent or original merchant), retrieving and analyzing the original payment fee
// then creating a ledger entry for the refunded amount if applicable.
func (s *RefundProcessor) CreditMDRToMerchant(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)
	merchantID := request.MerchantID // this var will decide who will got the refund charge

	if parentMerchantID != "" {
		merchantID = parentMerchantID
	}

	paymentFeeLedger, err := s.orchestratorSvc.FindByReference(ctx, request.PaymentID, constant.TypeFee)
	if err != nil {
		return err
	}
	if paymentFeeLedger == nil {
		return nil
	}
	// Check fee object
	paymentFeeLedgerMetadata := orchestratorModel.FeeTransactionMetadataObject{}
	_ = json.Unmarshal(paymentFeeLedger.AdditionalInfo.JSONText, &paymentFeeLedgerMetadata)

	// Product team, call it MDR
	paymentFeePercentage := 0.0
	if paymentFeeLedgerMetadata.AmountType != constant.MerchantFeeAmountType {
		paymentFeePercentage = paymentFeeLedgerMetadata.Percentage
	}

	// Value of payment fee that refunded to the merchant based on MDR percentage
	refundOfPaymentFee := 0.0
	if paymentFeePercentage > 0.0 {
		refundOfPaymentFee = (paymentFeePercentage / 100) * request.Amount
	}

	if refundOfPaymentFee > 0.0 {
		// Create refund fee (return the payment fee to the client)
		refundFeeLedgerID, _ := uuid.NewV7()
		refundFeeLedger := &orchestratorModel.CreateAccountTransactionRequest{
			UUID:                 refundFeeLedgerID,
			ReferenceID:          request.RefundID,
			Type:                 constant.TypeFeeRefund,
			MerchantID:           util.ParseUUID(merchantID),
			Currency:             paymentFeeLedger.Currency,
			Credit:               refundOfPaymentFee,
			Status:               constant.StatusSuccess,
			TransactionTimestamp: time.Now().UTC(),
			Usecase:              constant.TypePayment,
			AdditionalInfo: types.NullJSONText{
				Valid: true,
			},
		}
		refundFeeLedger.AdditionalInfo.JSONText, _ = json.Marshal(orchestratorModel.MetadataRefundOfPaymentFee{
			PaymentSessionID:   request.PaymentID,
			PaymentChargeID:    request.PaymentChargeID,
			PaymentFeeLedgerID: paymentFeeLedger.UUID.String(),
			FeeDetail: &feeModel.FeeMetadataObject{
				Type:        constant.TypeRefund,
				AmountType:  constant.MerchantFeePercentageType,
				Percentage:  paymentFeePercentage,
				FinalAmount: refundOfPaymentFee,
			},
		})

		if err = s.orchestratorSvc.PostAccountTransaction(ctx, refundFeeLedger); err != nil {
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	}
	return nil
}

// ChargeRefundFee processes the fee charging for refund transactions.
// It first determines the merchant ID to be charged (either the original merchant or parent merchant if available),
// then calculates the appropriate fee based on the refund parameters.
// If a fee applies, it creates a ledger transaction to record the fee charge.
func (s *RefundProcessor) ChargeRefundFee(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)
	merchantID := request.MerchantID // this var will decide who will got the refund charge

	if parentMerchantID != "" {
		merchantID = parentMerchantID
	}

	feeRefundAmount, feeRefundDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID:    merchantID,
		Reference:     constant.TypeRefund,
		PaymentMethod: "",
		ReferenceType: request.DestinationType,
	})

	if err != nil {
		s.logger.Error(ctx, "[RefundProcess] error when calculating fee", logger.Error(err))
	}

	if feeRefundAmount > 0 && feeRefundDetail != nil {
		// Create refund transfer fee (charge client because using the refund service)
		feeRefundLedgerID, _ := uuid.NewV7()
		feeRefundLedger := &orchestratorModel.CreateAccountTransactionRequest{
			UUID:                 feeRefundLedgerID,
			ReferenceID:          request.RefundID,
			Type:                 constant.TypeFee,
			MerchantID:           util.ParseUUID(merchantID),
			Currency:             constant.CurrencyIDR,
			Debit:                feeRefundAmount,
			Status:               constant.StatusSuccess,
			TransactionTimestamp: time.Now().UTC(),
			Usecase:              constant.TypePayment,
			AdditionalInfo: types.NullJSONText{
				Valid: true,
			},
		}
		feeRefundLedger.AdditionalInfo.JSONText, _ = json.Marshal(orchestratorModel.FeeTransactionMetadataObject{
			FeeMetadataObject: *feeRefundDetail,
		})

		if err = s.orchestratorSvc.PostAccountTransaction(ctx, feeRefundLedger); err != nil {
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	}
	return nil
}

func (s *RefundProcessor) updateRefundToSuccess(ctxTx context.Context, ctx context.Context, request *refundModel.RefundProcessRequest) error {
	// Update refund status to success
	request.Status = constant.RefundStatusSuccess
	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)

	// Update refund data
	if errUpdate := s.refundRepo.UpdateData(ctxTx, request.Refund); errUpdate != nil {
		return pkgErrs.New(response.HttpErrDatabase, errUpdate)
	}

	// Record refund success status
	s.refundSvc.RecordRefundStatusHistory(ctx, request.RefundID, constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess)

	// Update refund ledger status to success
	if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransactionByReferenceID(ctxTx, request.RefundLedgerReferenceID, constant.StatusSuccess, nil, nil); errUpdate != nil {
		s.logger.Error(ctx, "[RefundProcess] Update status account transaction (success)", logger.Error(errUpdate))
	}

	// charge submerchant as refund request to main-merchant
	if parentMerchantID != "" {
		err := s.ChargeSubMerchantToMerchant(ctxTx, request)
		if err != nil {
			s.logger.Error(ctxTx, "[RefundProcess] failed to charge Sub-Merchant", logger.Error(err))
			return pkgErrs.New(response.HttpErrInternal, err)
		}
	}

	if !constant.IsDirectPSP(request.PaymentMethodChannelType) {
		// Return the payment MDR fee (percentage of the merchant_fees) to the merchant
		// Ledger has been added in the previous step.

		// Calculate fee for the refund service
		err := s.ChargeRefundFee(ctxTx, request)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RefundProcessor) ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/ProcessUpdateBankTransferStatus")
	defer span.End()

	var (
		reasonType string
		reasonDesc string
	)

	refund, err :=
		s.getTransactionByExternalID(ctx, request.ExternalID)
	if err != nil {
		s.logger.Warn(ctx, "[ProcessUpdateBankTransferStatus] Cannot get refund transaction by external id",
			logger.Error(err), logger.String("externalID", request.ExternalID))
		return err
	}

	parentMerchantID, _ := ctx.Value(constant.CtxParentMerchantId).(string)
	if parentMerchantID == "" {
		merchant, err := s.merchantSvc.FindMerchantByID(ctx, refund.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "error when find merchant", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		if merchant.ParentID.Valid && merchant.ParentID.String != "" {
			ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
		}
	}

	// Find refund ledger by ID
	refundLedger, err := s.orchestratorSvc.FindByReference(ctx, refund.UUID, constant.TypeRefund)
	if err != nil {
		return err
	} else if refundLedger == nil {
		s.logger.Warn(ctx, "[ProcessUpdateBankTransferStatus] refund ledger is not found")
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	} else if refundLedger.Status != constant.StatusPending {
		s.logger.Warn(ctx, "[ProcessUpdateBankTransferStatus] refund ledger not in pending status")
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrRefundAlreadyProcessed)
	}

	refundOfPaymentFeeAmount := 0.0
	refundOfPaymentFee, err := s.orchestratorSvc.FindByReference(ctx, refund.UUID, constant.TypeFeeRefund)
	if err != nil {
		return err
	} else if refundOfPaymentFee != nil {
		refundOfPaymentFeeAmount = refundOfPaymentFee.Credit
	}

	refundLedgerAdditionalInfo := orchestratorModel.MetadataRefund{}
	_ = json.Unmarshal(refundLedger.AdditionalInfo.JSONText, &refundLedgerAdditionalInfo)

	// Find payment charge
	paymentCharge, err := s.orchestratorSvc.FindByID(ctx, refundLedgerAdditionalInfo.PaymentChargeID)
	if err != nil {
		return err
	} else if paymentCharge == nil {
		s.logger.Warn(ctx, "[ProcessUpdateBankTransferStatus] payment charge is not found")
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	paymentFee, err := s.orchestratorSvc.FindByReference(ctx, paymentCharge.ReferenceID, constant.TypeFee)
	if err != nil {
		return err
	} else if paymentFee == nil {
		s.logger.Warn(ctx, "[ProcessUpdateBankTransferStatus] payment fee is not found")
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	isDana := strings.Compare(request.ProcessorReference, constant.DanaPGProcessor) == 0
	// Request callback from Dana doesn't contain Amount
	// bypass validation amount for Dana
	if !isDana {
		// validation amount
		requestAmount := request.Amount.ToDecimal()
		refundAmount := decimal.NewFromFloat(refund.Amount)

		if requestAmount.Cmp(refundAmount) != 0 {
			s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Amount mismatch",
				logger.String("externalID", request.ExternalID))
			return constant.ErrInvalidRequestPayload
		}
	}

	processRefundRequest := &refundModel.RefundProcessRequest{
		RefundID:                 refund.UUID,
		Refund:                   refund,
		PaymentMethodType:        paymentCharge.Channel,
		PaymentProcessorID:       paymentCharge.ProcessorReferenceId,
		PaymentClientReferenceID: paymentCharge.ClientReferenceID,
		RefundLedgerID:           refundLedger.UUID.String(),
		RefundLedgerReferenceID:  refundLedger.ReferenceID,
		PaymentChargeID:          paymentCharge.UUID.String(),
		PaymentChargeAmount:      paymentCharge.Credit,
		PaymentFeeID:             paymentFee.UUID.String(),
		RefundOfPaymentFeeAmount: refundOfPaymentFeeAmount,
	}

	if paymentCharge.SettlementStatus.Valid {
		processRefundRequest.PaymentChargeSettlementStatus = paymentCharge.SettlementStatus.String
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

	if accountTransactionStatus == constant.StatusPending {
		s.logger.Info(ctx, "[ProcessUpdateBankTransferStatus] Refund processing is not processing yet")
		return nil
	}

	// Begin Tx for update disbursement status
	ctxTrx, trxErr := s.refundRepo.BeginTransaction(ctx)
	if trxErr != nil {
		return trxErr
	}

	txCompleted := false
	defer func() {
		if !txCompleted {
			if trxErr = s.refundRepo.RollbackTransaction(ctxTrx); trxErr != nil {
				return
			}
		}
	}()

	if accountTransactionStatus == constant.StatusFailed {
		// Update status refund to failed
		processRefundRequest.Status = constant.RefundStatusFailed
		if errUpdate := s.refundRepo.UpdateData(ctxTrx, processRefundRequest.Refund); errUpdate != nil {
			return pkgErrs.New(response.HttpErrDatabase, errUpdate)
		}

		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransactionByReferenceID(ctxTrx, processRefundRequest.RefundLedgerReferenceID, constant.StatusFailed, &reasonType, &reasonDesc); errUpdate != nil {
			s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Update status account transaction (failed)", logger.Error(errUpdate))
		}

		if err = s.refundRepo.CommitTransaction(ctxTrx); err != nil {
			s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Commit session transaction", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		txCompleted = true

		errCallback := s.refundSvc.SendCallback(ctx, processRefundRequest.UUID, processRefundRequest.MerchantID)
		if errCallback != nil {
			s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Send callback when failed ro process", logger.Error(errCallback))
		}

		return err
	}

	// Process SUCCESS refund
	if errUpdate := s.updateRefundToSuccess(ctxTrx, ctx, processRefundRequest); errUpdate != nil {
		s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Error when update refund to success", logger.Error(errUpdate))

		return errUpdate
	}

	// Commit Tx
	if trxErr = s.refundRepo.CommitTransaction(ctxTrx); trxErr != nil {
		return trxErr
	}
	txCompleted = true

	// Force settlement of unsettled payment
	if paymentCharge.SettlementStatus.String == constant.StatusPending {
		if errSettle := s.settlementSvc.ProcessSettlement(ctx, &settlementModel.ProcessSettlementRequest{
			MerchantID:           processRefundRequest.MerchantID,
			Type:                 constant.SettlementTransaction,
			TransactionID:        processRefundRequest.PaymentChargeID,
			FeeTransactionID:     processRefundRequest.PaymentFeeID,
			ByPassSettlementHold: true,
		}); errSettle != nil {
			s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] failed to force settlement", logger.Error(err))
		}
	}

	errCallback := s.refundSvc.SendCallback(ctx, processRefundRequest.UUID, processRefundRequest.MerchantID)
	if errCallback != nil {
		s.logger.Error(ctx, "[ProcessUpdateBankTransferStatus] Send callback on final state", logger.Error(errCallback))
	}

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

func (s *RefundProcessor) handlingReasonTypeAndDescFromTransfer(
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

func (s *RefundProcessor) getTransactionByExternalID(ctx context.Context, externalID string) (*refundModel.Refund, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/getTransactionByExternalID")
	defer segment.End()

	existedTransaction, err := s.orchestratorSvc.FindByID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	refundID := existedTransaction.ReferenceID

	// Find disbursement by ID
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, err
	} else if refund == nil {
		err = constant.ErrRefundNotFound
		return nil, err
	}

	return refund, nil
}
