package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"google.golang.org/protobuf/proto"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *UnifiedPaymentService) ProcessNotification(ctx context.Context, request *unifiedPaymentModel.PaymentNotificationRequest) (err error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ProcessNotification")
	defer span.End()

	s.logger.Info(ctx, "[UnifiedPaymentV2] Process payment notification", logger.Any("request", request))
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(ctx, "Panic recovery from ProcessNotification", logger.Error(fmt.Errorf("%v", r)))
		}

		s.logger.Info(ctx, "[UnifiedPaymentV2] Finish payment notification")
	}()

	paymentSession, err := s.paymentSvc.GetDetailByID(ctx, request.PaymentSessionID)
	if err != nil {
		return err
	}
	unifiedPaymentResp := paymentSession.ToUnifiedPaymentResponse()

	if !unifiedPaymentResp.IsStaticPayment() {
		lockKey := fmt.Sprintf(constant.LockKeyUnifiedPaymentNotificationKey, paymentSession.UUID, request.ChargeStatus)
		isLockAcquired, err := s.redis.SetNX(ctx, lockKey, "1", constant.LockKeyUnifiedPaymentNotificationExpiry).Result()
		if err != nil {
			s.logger.Error(ctx, "error when acquire unified payment notification lock", logger.Error(err), logger.String("key", lockKey))
			return pkgErrors.New(response.HttpErrInternal, constant.ErrAcquireLockUnifiedPaymentNotification)
		}
		if !isLockAcquired {
			s.logger.Info(ctx, "failed acquire unified payment notification lock", logger.String("key", lockKey))
			return pkgErrors.New(response.HttpErrInternal, constant.ErrAcquireLockUnifiedPaymentNotification)
		}
		s.logger.Info(ctx, "acquire unified payment notification lock", logger.String("key", lockKey))
	}

	// Set derived merchant ID in context for sub-merchant transactions.
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, paymentSession.MerchantID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, constant.ErrFindMerchant)
	}
	if merchant != nil && merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
		ctx = context.WithValue(ctx, constant.CtxDerivedMerchantID, merchant.ParentID.String)
	}

	if err = s.validatePaymentFinalStatus(ctx, paymentSession.UUID, paymentSession.Status); err != nil {

		currentStatus, attemptedStatusChange := paymentSession.Status, request.ChargeStatus

		if currentStatus == constant.UnifiedPaymentSessionStatusPaid && attemptedStatusChange == constant.ChargeStatusSuccess {
			s.logger.Info(ctx, "ignore success payment notif on paid payment session")
			return err
		}

		if currentStatus == constant.UnifiedPaymentSessionStatusCancelled && attemptedStatusChange == constant.ChargeStatusFailed {
			// known condition is caused from expiry payment handling: https://app.clickup.com/t/86d02fhtz
			// regardless this is not a status change that we should worry about, as changing between cancelled and failed does not affects transaction
			// so exit early without sending slack alert
			return err
		}

		// Send Slack alert for payment final status conflict
		s.sendPaymentFinalStatusConflictAlert(
			ctx,
			paymentSession.UUID,
			currentStatus,
			attemptedStatusChange,
			request.ChargeID,
			request.Processor,
			request.PaymentMethodType,
		)
		return err
	}

	// Prepare recurring payment detail data.
	if unifiedPaymentResp.RecurringID != "" {
		paymentSession.RecurringPayment = &unifiedPaymentModel.MetadataRecurringPayment{
			InitiateFirstAuthorization: unifiedPaymentResp.InitiateFirstAuthorization,
			FirstAuthorizationMethod:   unifiedPaymentResp.FirstAuthorizationMethod,
			FirstAuthorizationOrderID:  unifiedPaymentResp.FirstAuthorizationOrderID,
			BillingCycle:               unifiedPaymentResp.RecurringBillingCycle,
		}
	}
	// Release the exclusive lock for recurring payment transactions with a finalized status.
	isRecurringPaymentFinalStatus := unifiedPaymentResp.RecurringID != "" &&
		(request.ChargeStatus == constant.ChargeStatusSuccess || request.ChargeStatus == constant.ChargeStatusFailed)
	if isRecurringPaymentFinalStatus {
		// Process executed after successful payment notification handling (scheduled).
		defer func() {
			if err != nil {
				return
			}
			recurringPaymentType := constant.RecurringPaymentTypeSubsequentPayment
			if unifiedPaymentResp.InitiateFirstAuthorization {
				recurringPaymentType = constant.RecurringPaymentTypeFirstAuthorization
			}
			s.logger.Info(ctx, "Exclusive lock released for recurring ID "+unifiedPaymentResp.RecurringID+" type "+recurringPaymentType)

			processKey := fmt.Sprintf(constant.RecurringPaymentMutualExclusionKey, recurringPaymentType, unifiedPaymentResp.RecurringID)
			if delErr := s.redis.Del(ctx, processKey).Err(); delErr != nil {
				s.logger.Error(ctx, "Failed to release recurring payment exclusive lock", logger.Error(delErr))
			}
		}()
	}

	defer s.handleAutoSplitPayment(ctx, request, paymentSession, err)

	// Set specific usecase attribute
	paymentSession.CardFundedPayout = unifiedPaymentResp.CardFundedPayout
	paymentSession.AutoSplitPayment = unifiedPaymentResp.AutoSplitPayment

	switch request.ChargeStatus {
	case constant.ChargeStatusSuccess:
		s.RecordChargeStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorProcessor, constant.ChargeStatusHistorySuccess)

		if unifiedPaymentResp.IsStaticPayment() {
			return s.payStaticPaymentCharge(ctx, paymentSession, request)
		}
		return s.payCharge(ctx, paymentSession, request)

	case constant.ChargeStatusFailed:
		// Tech Debt: Mark subsequent transactions as CANCELED when the initial card-funded payout transaction fails.
		s.RecordChargeStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorProcessor, constant.ChargeStatusHistoryFailed)
		return s.handleFailedCharge(ctx, paymentSession, request)

	case constant.ChargeStatusProcessing, constant.ChargeStatusWaitingForCapture:
		s.RecordChargeStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorProcessor, constant.ChargeStatusHistoryProcessing)
		return s.handleProcessCharge(ctx, paymentSession, request)

	case constant.ChargeStatusWaitingForAuthentication:
		// No need to record status history
		return s.handleWaitingForAuthentication(ctx, paymentSession)

	default:
		return constant.ErrStatusNotAllowed
	}
}

func (s *UnifiedPaymentService) handleFailedCharge(ctx context.Context, payment *paymentModel.Payment, request *unifiedPaymentModel.PaymentNotificationRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/handleFailedCharge")
	defer span.End()

	s.logger.Info(ctx, "[UnifiedPaymentV2] Processing failed payment", logger.Any("payment_id", payment.UUID), logger.Any("charge_id", request.ChargeID))

	// defer for nonblocking FDS update
	// executed after the transaction is  commited
	defer func() {
		// Auto-update FDS for failed CC transactions
		if payment.PaymentMethod.Type == constant.ChannelCreditCard {
			// new context to avoid any cancelation
			bgCtx := context.Background()
			updateResp, err := s.fdsSvc.UpdateTransaction(bgCtx, request.ChargeID, nil)
			if err != nil {
				s.logger.Error(bgCtx, "Failed to auto-update FraudNet for CC failure", logger.Error(err))
			} else {
				s.logger.Info(bgCtx, "FDS Update Response", logger.Any("response", updateResp))
			}
		}
	}()

	// Start a transaction to update all related records
	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	// Update payment status to cancelled (for failed payments)
	payment.Status = constant.UnifiedPaymentSessionStatusCancelled
	payment.UpdatedAt = time.Now().UTC()
	err := s.paymentRepo.UpdatePaymentStatus(ctxTrx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		s.logger.Error(ctx, "Failed to update payment status", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Update transaction metadata and status
	paymentLedger, err := s.accountTransactionRepo.FindByID(ctxTrx, request.ChargeID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find account transaction", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// If we have a payment ledger record, update its status
	if paymentLedger != nil {
		paidAmount := commonModel.Amount{
			Currency: request.Amount.Currency,
			Value:    decimal.NewFromFloat(request.Amount.Value).StringFixed(2),
		}

		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorReferenceId:   request.ProcessorID,
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			Status:                 constant.StatusFailed,
			Channel:                request.PaymentMethodType,
			Amount:                 paidAmount,
			FailureCode:            s.GetFailureCodeOfMethodDetail(constant.StatusFailed, request.ChargePaymentMethodDetails),
		}

		if request.ChargePaymentMethodDetails != nil {
			updateRequest.MethodDetail = request.ChargePaymentMethodDetails
		}

		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}

		if err := s.paymentSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			s.logger.Error(ctx, "Failed to update pending ledger", logger.Error(err))
			return err
		}
	}

	if payment.CardFundedPayout != nil && payment.CardFundedPayout.Sequence == 1 {
		_ = s.cardFundedPayoutSvc.ProcessInitialCardFundedPayoutAuthFailure(ctxTrx, payment.MerchantID, *payment.ReferenceID)
		// Send a status notification to close the status check pop-up.
		s.sendStompNotification(ctx, payment, util.ValueOfPtr(payment.ReferenceID))
	}

	// Commit the transaction
	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		s.logger.Error(ctx, "Failed to commit transaction", logger.Error(errCommit))
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true

	// Send callback for failed charge
	s.SendCallback(ctx, payment)

	return nil
}

func (s *UnifiedPaymentService) handleWaitingForAuthentication(ctx context.Context, payment *paymentModel.Payment) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/handleWaitingForAuthentication")
	defer span.End()

	s.logger.Info(ctx, "[UnifiedPaymentV2] Processing waiting for authentication payment", logger.String("paymentId", payment.UUID))

	if payment.PaymentMethod.Type != constant.ChannelCreditCard {
		s.logger.Info(ctx, "skip process for non card transaction", logger.Any("paymentId", payment.UUID))
		return nil
	}

	unifiedPaymentMetadata := payment.ToUnifiedPaymentMetadata()
	if unifiedPaymentMetadata == nil {
		s.logger.Info(ctx, "empty unifiedPaymentMetadata", logger.Any("paymentId", payment.UUID))
		return nil
	}

	if unifiedPaymentMetadata.Mode != constant.UnifiedPaymentModeAPI {
		s.logger.Info(ctx, "do not process for non API mode", logger.Any("paymentId", payment.UUID))
		return nil
	}

	paymentDTO := payment.ToDTO()

	// Set retryable confirmation for waiting for authentication event
	isRetryable := true
	unifiedPaymentMetadata.RetryableConfirmation = &isRetryable

	// Set metadata
	if metaDataB, errMarshal := json.Marshal(unifiedPaymentMetadata); errMarshal != nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrInvalidMarshalData)
	} else if metaDataB != nil {
		metadataStr := string(metaDataB)
		paymentDTO.Metadata = &metadataStr
	}

	if err := s.paymentRepo.UpdatePaymentData(ctx, paymentDTO); err != nil {
		s.logger.Error(ctx, "Failed to update payment data", logger.Any("paymentId", payment.UUID), logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *UnifiedPaymentService) payCharge(ctx context.Context, payment *paymentModel.Payment, request *unifiedPaymentModel.PaymentNotificationRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/payCharge")
	defer span.End()

	if payment.Type != constant.PaymentTypeCardFundedPayout && request.Amount.Value > 0 && request.Amount.Value != payment.Amount.InexactFloat64() {
		s.logger.Error(ctx, "error payment amount not match with paid amount", logger.Error(constant.ErrPaymentAmountNotMatch))
		return pkgErrors.New(response.HttpStatusErrorUnprocessableContent, constant.ErrPaymentAmountNotMatch)
	}

	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	// Modification of special payment channels for payments via credit cards
	if payment.PaymentMethod.Type == constant.ChannelCreditCard && payment.Type != constant.PaymentTypeCardFundedPayout {
		merchant, err := s.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
			return err
		}
		channel := "FOREIGN_"
		if merchant.BusinessCountry.String == request.ChargePaymentMethodDetails.Card.BinInformations.Country {
			channel = "LOCAL_"
		}
		channel += strings.ToUpper(request.ChargePaymentMethodDetails.Card.BinInformations.Brand)

		payment.PaymentMethod.Acquirer = channel // For calculation payment fee per channel

		cardAcquirer, err := s.GetCardMIDAcquirer(ctx, payment, request)
		if err != nil {
			return err
		}

		if request.ChargePaymentMethodDetails.Card.MIDInfo != nil {
			request.ChargePaymentMethodDetails.Card.MIDInfo.Acquirer = cardAcquirer
		}

	}

	err := s.changeSettlementModelForCardPayment(ctxTrx, payment, request)
	if err != nil {
		return err
	}

	if err := s.paymentSvc.DeterminePaymentFee(&ctxTrx, payment); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	paymentLedgerStatus := constant.StatusSuccess
	payment.Status = constant.UnifiedPaymentSessionStatusPaid
	if payment.IsAutoSplitPaymentAuth() {
		paymentLedgerStatus = constant.StatusPending
		payment.Status = constant.UnifiedPaymentSessionStatusProcessing
	}
	payment.UpdatedAt = time.Now().UTC()
	err = s.paymentRepo.UpdatePaymentStatus(ctxTrx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if !request.TrxDatetime.IsZero() {
		payment.TrxDatetime = &request.TrxDatetime
	} else {
		request.TrxDatetime = time.Now().UTC()
	}
	payment.ProcessorTransactionID = request.ProcessorTransactionID
	payment.Processor = request.Processor
	payment.ProcessorID = request.ProcessorID
	payment.TrxDatetime = &request.TrxDatetime

	if payment.ProcessorReferenceNumber != nil {
		payment.ReconReferenceNo = *payment.ProcessorReferenceNumber
	}

	if request.PaymentMethodType == constant.UnifiedPaymentMethodCard &&
		request.ChargePaymentMethodDetails != nil && request.ChargePaymentMethodDetails.Card != nil {

		payment.ReconReferenceNo = util.ValueOfPtr(request.ChargePaymentMethodDetails.Card.AuthorizationResult).AuthorizationID

		if util.ValueOfPtr(request.ChargePaymentMethodDetails.Card.SaveForFutureUse) {
			customerUseCase := payment.GetOneDollarAuthorizationUseCase()
			if err = s.storeFutureUseOfCustomerPaymentMethodCard(ctx, payment.CustomerID, customerUseCase, request.ChargePaymentMethodDetails.Card); err != nil {
				return err
			}
		}
	}
	paymentLedger, err := s.accountTransactionRepo.FindByID(ctxTrx, request.ChargeID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	paidAmount := commonModel.Amount{
		Currency: request.Amount.Currency,
		Value:    decimal.NewFromFloat(request.Amount.Value).StringFixed(2),
	}
	if paymentLedger == nil {
		if err := s.paymentSvc.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
			Status:  paymentLedgerStatus,
			Channel: request.PaymentMethodType,
			Amount:  paidAmount,
		}); err != nil {
			return err
		}

	} else {
		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorReferenceId:   request.ProcessorID,
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			TrxDatetime:            payment.TrxDatetime,
			Status:                 paymentLedgerStatus,
			Channel:                request.PaymentMethodType,
			Amount:                 paidAmount,
		}
		if request.ChargePaymentMethodDetails != nil {
			updateRequest.MethodDetail = s.updateVAMethodDetailOnSuccess(paymentLedger, request.ChargePaymentMethodDetails)
		}
		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}
		if err := s.paymentSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			return err
		}
	}

	if payment.RecurringPayment != nil {
		recurringPaymentRequest := recurringContractModel.UpdateRecurringPaymentRequest{
			MerchantID:       payment.MerchantID,
			RecurringID:      util.ValueOfPtr(payment.RecurringContractID),
			TransactionID:    paymentLedger.UUID.String(),
			PaymentMethodID:  payment.PaymentMethodID,
			RecurringPayment: payment.RecurringPayment,
			UpdatedBy:        util.ValueOfPtr(payment.CreatedBy),
		}
		if request.ChargePaymentMethodDetails != nil && request.ChargePaymentMethodDetails.Card != nil {
			recurringPaymentRequest.PaymentTokenID = request.ChargePaymentMethodDetails.Card.Token
		}
		if err := s.recurringContractSvc.UpdateRecurringPayment(ctxTrx, recurringPaymentRequest); err != nil {
			return err
		}
	}

	if payment.CardFundedPayout != nil && payment.CardFundedPayout.Sequence == 1 {
		_ = s.cardFundedPayoutSvc.ProcessPendingSubsequentPayments(ctx, payment.MerchantID, *payment.ReferenceID)
		// Send a status notification to close the status check pop-up.
		s.sendStompNotification(ctx, payment, util.ValueOfPtr(payment.ReferenceID))
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true

	// Autosplit payment proceed the parent payment with processing state after 3ds step
	// should not send the processing callback
	if payment.IsAutoSplitPaymentAuth() {
		return nil
	}

	// Send callback on paid charge
	s.SendCallback(ctx, payment)

	return nil
}

func (s *UnifiedPaymentService) payStaticPaymentCharge(ctx context.Context, payment *paymentModel.Payment, request *unifiedPaymentModel.PaymentNotificationRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/payStaticPaymentCharge")
	defer span.End()

	// Validation for static closed
	if payment.Amount.InexactFloat64() > 0 && request.Amount.Value != payment.Amount.InexactFloat64() {
		s.logger.Error(ctx, "error payment amount not match with paid amount", logger.Error(constant.ErrPaymentAmountNotMatch))
		return pkgErrors.New(response.HttpStatusErrorUnprocessableContent, constant.ErrPaymentAmountNotMatch)
	}

	// Set distributed lock
	mutex := s.redis.NewMutex(
		"backend-portal:static-payment:"+payment.UUID+":pay-charge:lock",
		redsync.WithExpiry(120*time.Second),
		redsync.WithRetryDelay(10*time.Millisecond),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
	if err := mutex.LockContext(ctx); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}
	defer func() {
		if _, err := mutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed unlock process", logger.Error(err))
		}
	}()

	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	payment.Amount = decimal.NewFromFloat(request.Amount.Value)
	if err := s.paymentSvc.DeterminePaymentFee(&ctxTrx, payment); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if !request.TrxDatetime.IsZero() {
		payment.TrxDatetime = &request.TrxDatetime
	} else {
		request.TrxDatetime = time.Now().UTC()
	}
	payment.ProcessorTransactionID = request.ProcessorTransactionID
	payment.Processor = request.Processor
	payment.ProcessorID = request.ProcessorID
	payment.TrxDatetime = &request.TrxDatetime
	if request.ChargePaymentMethodDetails != nil && request.ChargePaymentMethodDetails.VirtualAccount != nil {
		payment.BankReferenceId = request.ChargePaymentMethodDetails.VirtualAccount.BankReferenceNo
	}

	if payment.ProcessorReferenceNumber != nil {
		payment.ReconReferenceNo = *payment.ProcessorReferenceNumber
	}

	chargeID, _ := uuid.NewV7()
	if err := s.paymentSvc.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
		Status:  constant.StatusSuccess,
		Channel: request.PaymentMethodType,
		Amount: commonModel.Amount{
			Value:    fmt.Sprintf("%.2f", request.Amount.Value),
			Currency: request.Amount.Currency,
		},
		ChargeID:     chargeID.String(),
		ChargeStatus: constant.ChargeStatusSuccess,
	}); err != nil {
		return err
	}

	// Update metadata for static payment
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	summaryTransaction, err := s.summaryStaticPaymentTransaction(
		ctxTrx,
		unifiedPaymentMetadata.SummaryTransaction,
		payment.UUID,
		payment.MerchantID,
		request.Amount.Value,
	)
	if err != nil {
		return err
	}

	if err := s.paymentRepo.UpdatePaymentMetadataById(ctxTrx, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
		SummaryTransaction: summaryTransaction,
	}); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	// Send callback on paid charge
	s.sendPaymentChargeCallback(ctx, chargeID.String(), payment)

	return nil
}

func (s *UnifiedPaymentService) summaryStaticPaymentTransaction(
	ctx context.Context, summaryTransaction *unifiedPaymentModel.SummaryTransaction, paymentID, merchantID string, paidAmount float64) (*unifiedPaymentModel.SummaryTransaction, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/summaryStaticPaymentTransaction")
	defer span.End()

	if summaryTransaction == nil || summaryTransaction.CountPaidAmount <= 0 {
		merchantUUID, err := uuid.Parse(merchantID)
		if err != nil {
			s.logger.Error(ctx, "err parsing merchant id", logger.Error(err))
			return nil, pkgErrors.New(response.HttpErrInternal, err)
		}

		aggregateTransaction, err := s.accountTransactionRepo.GetAggregateTransactionByReference(ctx, &orchestrator_model.GetSummaryTransactionByReferenceRequest{
			MerchantID:    merchantUUID,
			ReferenceType: constant.TypePayment,
			ReferenceID:   paymentID,
			Status:        constant.StatusSuccess,
		})
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		}

		summaryTransaction = &unifiedPaymentModel.SummaryTransaction{
			SumPaidAmount:   aggregateTransaction.SumOfCredit,
			CountPaidAmount: aggregateTransaction.CountOfCredit,
		}

		return summaryTransaction, nil
	}

	summaryTransaction.CountPaidAmount += 1
	summaryTransaction.SumPaidAmount += paidAmount
	return summaryTransaction, nil
}

func (s *UnifiedPaymentService) handleProcessCharge(ctx context.Context, payment *paymentModel.Payment, request *unifiedPaymentModel.PaymentNotificationRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/handleProcessCharge")
	defer span.End()

	if payment.Status != constant.UnifiedPaymentSessionStatusRequireAction &&
		!(payment.Status == constant.UnifiedPaymentSessionStatusProcessing && request.ChargeStatus == constant.ChargeStatusWaitingForCapture) {
		s.logger.Info(ctx, "payment is not in processable state, should ignore the notification", logger.String("paymentID", payment.UUID))
		return nil
	}

	s.logger.Info(ctx, "[UnifiedPaymentV2] payment is processed", logger.Any("payment_id", payment.UUID), logger.Any("charge_id", request.ChargeID))

	// Start a transaction to update all related records
	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	payment.Status = constant.UnifiedPaymentSessionStatusProcessing
	payment.UpdatedAt = time.Now().UTC()
	err := s.paymentRepo.UpdatePaymentStatus(ctxTrx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		s.logger.Error(ctx, "Failed to update payment status", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	paymentLedger, err := s.accountTransactionRepo.FindByID(ctxTrx, request.ChargeID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find account transaction", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentLedger != nil {
		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorReferenceId:   request.ProcessorID,
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			Status:                 constant.StatusPending,
			Channel:                request.PaymentMethodType,
			ChargeStatus:           request.ChargeStatus,
		}

		if request.ChargePaymentMethodDetails != nil {
			updateRequest.MethodDetail = request.ChargePaymentMethodDetails
		}

		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}

		if err := s.paymentSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			s.logger.Error(ctx, "Failed to update pending ledger", logger.Error(err))
			return err
		}
	}

	if request.ChargeStatus == constant.ChargeStatusWaitingForCapture {
		if err := s.changeSettlementModelForCardPayment(ctxTrx, payment, request); err != nil {
			return err
		}
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		s.logger.Error(ctx, "Failed to commit transaction", logger.Error(errCommit))
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true

	// Extend payment expiration
	if request.ChargeStatus == constant.ChargeStatusProcessing {
		evalCtx := ffcontext.NewEvaluationContext(s.config.Environment)
		evalCtx.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)
		evalCtx.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, payment.MerchantID)
		delayDuration, _ := ffclient.IntVariation(constant.FeatureFlagExpireProcessedPaymentInMinute, evalCtx, 15)
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:         payment.UUID,
				MerchantID:   payment.MerchantID,
				Note:         fmt.Sprintf("Payment already processed at %v (UTC)", time.Now()),
				ChargeStatus: constant.ChargeStatusProcessing,
			},
			time.Duration(delayDuration)*time.Minute,
		); err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
		}
	} else if request.ChargeStatus == constant.ChargeStatusWaitingForCapture {
		// Save card future use for authorization flow
		if request.PaymentMethodType == constant.UnifiedPaymentMethodCard &&
			request.ChargePaymentMethodDetails != nil && request.ChargePaymentMethodDetails.Card != nil {

			payment.ReconReferenceNo = util.ValueOfPtr(request.ChargePaymentMethodDetails.Card.AuthorizationResult).AuthorizationID

			if util.ValueOfPtr(request.ChargePaymentMethodDetails.Card.SaveForFutureUse) {
				customerUseCase := payment.GetOneDollarAuthorizationUseCase()
				if err = s.storeFutureUseOfCustomerPaymentMethodCard(ctx, payment.CustomerID, customerUseCase, request.ChargePaymentMethodDetails.Card); err != nil {
					return err
				}
			}
		}

		delayDuration := time.Until(util.ValueOfPtr(payment.ExpiredAt))
		if delayDuration > time.Duration(s.config.UnifiedPaymentConfig.MaxAuthorizeTransactionMinutes)*time.Minute {
			delayDuration = time.Duration(s.config.UnifiedPaymentConfig.MaxAuthorizeTransactionMinutes) * time.Minute
		}
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:         payment.UUID,
				MerchantID:   payment.MerchantID,
				ChargeStatus: constant.ChargeStatusWaitingForCapture,
			},
			delayDuration,
		); err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
		}
	}

	s.SendCallback(ctx, payment)

	return nil
}

func (s *UnifiedPaymentService) storeFutureUseOfCustomerPaymentMethodCard(
	ctx context.Context, customerID, customerUseCase string, cardData *unifiedPaymentModel.ChargePaymentMethodDetailCard) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/storeFutureUseOfCustomerPaymentMethod")
	defer span.End()

	if customerID == "" {
		return nil
	}

	customer, err := s.customerRepo.FindCustomerById(ctx, customerID)
	if err != nil {
		s.logger.Error(ctx, "err finding customer id", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if customer == nil {
		s.logger.Info(ctx, "customer not found")
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrCustomerNotFound)
	}

	var customerPaymentMethods []*unifiedPaymentModel.CustomerPaymentMethod
	paymentMethodList, exist := customer.Metadata["paymentMethods"]
	if exist {
		customerPaymentMethods, _ = util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethod](paymentMethodList)
	}
	for _, paymentMethod := range customerPaymentMethods {
		if paymentMethod.Card != nil && paymentMethod.Card.Fingerprint == cardData.Fingerprint {
			cardData.Token = paymentMethod.Token
			s.logger.Info(ctx, "fingerprint already stored in customer data", logger.String("customerID", customerID))
			return nil
		}
	}

	// If fingerprint is not stored
	customer.PhoneNumber = customer.OriginalPhoneNumber

	cardHolderFirstName := ""
	cardHolderLastName := ""
	if cardData.CardHolderName != "" {
		cardHolderNameParts := strings.Fields(cardData.CardHolderName)
		cardHolderFirstName = cardHolderNameParts[0]
		if len(cardHolderNameParts) > 1 {
			cardHolderLastName = strings.Join(cardHolderNameParts[1:], " ")
		}
	}

	cardOrigin := constant.CardOriginLocal
	if cardData.BinInformations.IsForeign() {
		cardOrigin = constant.CardOriginForeign
	}

	cardData.Token = uuid.NewString()
	customerPaymentMethods = append([]*unifiedPaymentModel.CustomerPaymentMethod{
		{
			Token:          cardData.Token,
			PaymentMethod:  constant.UnifiedPaymentMethodCard,
			PaymentChannel: cardData.BinInformations.Brand,
			Status:         constant.StoredPaymentMethodStatusActive,
			CreatedAt:      time.Now().UTC(),
			Card: &unifiedPaymentModel.CustomerPaymentMethodCard{
				Fingerprint:         cardData.Fingerprint,
				Network:             cardData.BinInformations.Brand,
				First6:              cardData.First6,
				First8:              cardData.First8,
				Last4:               cardData.Last4,
				ExpMonth:            cardData.ExpMonth.String(),
				ExpYear:             cardData.ExpYear.String(),
				CardHolderFirstName: cardHolderFirstName,
				CardHolderLastName:  cardHolderLastName,
				CardName:            cardData.CardName,
				IssuingBank:         cardData.BinInformations.IssuingBank,
				CardOrigin:          cardOrigin,
			},
		},
	},
		customerPaymentMethods...,
	)

	if customerUseCase != "" {
		customer.Metadata["useCase"] = customerUseCase
	}

	customer.Metadata["paymentMethods"] = customerPaymentMethods
	if err = s.customerRepo.Update(ctx, customer); err != nil {
		s.logger.Error(ctx, "err updating customer id", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}

// updateVAMethodDetailOnSuccess updates the VirtualAccountName field of the provided methodDetail
// with the value from the paymentLedger's AdditionalInfo, if available and non-empty.
// It returns the updated methodDetail. If either methodDetail or paymentLedger is nil,
// or if the VirtualAccount field is nil, the original methodDetail is returned unchanged.
// TODO: remove the func when the snap core processor also publish virtual account name
func (s *UnifiedPaymentService) updateVAMethodDetailOnSuccess(paymentLedger *orchestrator_model.AccountTransactionWithUseCase, methodDetail *unifiedPaymentModel.ChargePaymentMethodDetails) *unifiedPaymentModel.ChargePaymentMethodDetails {
	if methodDetail == nil || paymentLedger == nil {
		return methodDetail
	}

	if methodDetail.VirtualAccount == nil {
		return methodDetail
	}

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(paymentLedger.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	if chargeMethodDetails.VirtualAccount == nil {
		return methodDetail
	}

	if chargeMethodDetails.VirtualAccount.VirtualAccountName != "" {
		methodDetail.VirtualAccount.VirtualAccountName = chargeMethodDetails.VirtualAccount.VirtualAccountName
	}

	return methodDetail
}

// We need to change the settlement model of card payment before the ledger updated
// because card have multiple MID that may have different settlement model.
func (s *UnifiedPaymentService) changeSettlementModelForCardPayment(ctx context.Context, payment *paymentModel.Payment, notificationReq *unifiedPaymentModel.PaymentNotificationRequest) error {
	if payment == nil || notificationReq == nil {
		return nil
	}

	if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		return nil
	}

	if notificationReq.ChargePaymentMethodDetails == nil || notificationReq.ChargePaymentMethodDetails.Card == nil {
		return nil
	}

	if notificationReq.ChargePaymentMethodDetails.Card.MIDInfo == nil {
		return nil
	}

	midType := notificationReq.ChargePaymentMethodDetails.Card.MIDInfo.Type
	// Force to use AGGREGATOR for CFP
	if payment.Type == constant.PaymentTypeCardFundedPayout {
		midType = constant.PaymentMethodChannelTypeAggregator
	}

	paymentMethodChannelType := constant.PaymentMethodChannelTypeAggregator

	paymentLedger, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment ledger", logger.Error(err), logger.String("paymentID", payment.UUID))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentLedger == nil {
		s.logger.Warn(ctx, "payment ledger not found", logger.String("paymentID", payment.UUID))
		return nil
	}

	if paymentLedger.SettlementModel.Valid {
		paymentMethodChannelType = paymentLedger.SettlementModel.String
	}

	if constant.ChannelTypeToMidType(paymentMethodChannelType) == midType {
		s.logger.Info(ctx, "settlement model already equal don't need to update it", logger.String("paymentID", payment.UUID), logger.String("settlement model", paymentMethodChannelType))
		return nil
	}
	s.logger.Info(ctx, "settlement model should changed",
		logger.String("paymentID", payment.UUID),
		logger.String("settlement model", paymentMethodChannelType),
		logger.String("midType", midType),
	)

	err = s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(ctx, orchestrator_model.UpdatePaymentTransactionRequest{
		SettlementModel: util.ValueToPtr(constant.MidTypeToChannelType(midType)),
		LedgerId:        paymentLedger.UUID.String(),
		UpdatedAt:       time.Now(),
	}, orchestrator_model.MetadataPayment[any]{})

	if errors.Is(err, constant.ErrDataNotFound) {
		return pkgErrors.New(response.HttpErrUnprocessableContent, err)
	}

	if err != nil {
		s.logger.Error(ctx, "failed to update settlement model", logger.Error(err), logger.String("paymentID", payment.UUID))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *UnifiedPaymentService) sendPaymentFinalStatusConflictAlert(ctx context.Context, paymentID, previousStatus, afterStatus, chargeID, processor, paymentMethodType string) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/sendPaymentFinalStatusConflictAlert")
	defer span.End()

	// Return if not final status
	if !slices.Contains([]string{constant.ChargeStatusSuccess, constant.ChargeStatusFailed}, afterStatus) {
		return
	}

	// Extract trace ID from context
	traceID := ""
	if id, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); ok && id != "" {
		traceID = id
	}

	fields := []*slackPb.AttachmentField{
		{Title: "Payment ID", Value: paymentID, Short: true},
		{Title: "Charge ID", Value: chargeID, Short: true},
		{Title: "Previous Payment Status", Value: previousStatus, Short: true},
		{Title: "Attempted Charge Status", Value: afterStatus, Short: true},
		{Title: "Payment Method", Value: paymentMethodType, Short: true},
		{Title: "Processor", Value: processor, Short: true},
		{Title: "Trace ID", Value: traceID, Short: true},
		{Title: "Timestamp", Value: time.Now().UTC().Format(time.RFC3339), Short: true},
		{Title: "Error", Value: "Payment status already in final status - cannot process notification", Short: false},
	}

	slackMessage := &slackPb.PostWebhookCmd{
		URL:    s.config.SlackConfig.PGAlertWebHookURL,
		Color:  slackPb.Color_GOOD,
		Title:  "<!subteam^S06L0CMNUQH> <!subteam^S070CMVB03C> Payment Final Status Conflict Detected",
		Fields: fields,
	}

	rawSlackMessage, err := proto.Marshal(slackMessage)
	if err != nil {
		s.logger.Error(ctx, "Failed to marshal slack message", logger.Error(err))
		return
	}

	if err := s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage); err != nil {
		s.logger.Error(ctx, "Failed to publish slack alert", logger.Error(err))
	}
}

func (s *UnifiedPaymentService) validatePaymentFinalStatus(ctx context.Context, paymentID, status string) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/validatePaymentFinalStatus")
	defer span.End()

	if slices.Contains([]string{
		constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedPaymentSessionStatusCancelled,
		//constant.UnifiedPaymentSessionStatusExpired, // status EXPIRED still can be replaced
	}, status) {
		s.logger.Error(ctx, "payment status already in final status", logger.String("paymentID", paymentID))
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentAlreadyInFinalStatus)
	}

	return nil
}

func (s *UnifiedPaymentService) handleAutoSplitPayment(ctx context.Context, request *unifiedPaymentModel.PaymentNotificationRequest, paymentSession *paymentModel.Payment, err error) {
	if err != nil {
		return
	}

	metadata := paymentSession.ToUnifiedPaymentMetadata()
	// skip non-auto split payment
	if metadata == nil || metadata.AutoSplitPayment == nil {
		return
	}

	paymentSession.AutoSplitPayment = metadata.AutoSplitPayment
	if paymentSession.IsAutoSplitPaymentAuth() {
		if request.ChargeStatus != constant.ChargeStatusSuccess {
			return
		}
		err := s.internalUnifiedPaymentSvc.InitiateSplitPayment(ctx, &paymentModel.ProcessSplitPaymentRequest{
			ParentPaymentID:   paymentSession.UUID,
			ThreeDSCallbackID: request.GetCardThreeDSCallbackID(),
			FingerprintID:     request.GetCardFingerprintID(),
		})
		if err != nil {
			s.logger.Error(ctx, "Failed to process split payment",
				logger.Error(err),
				logger.String("paymentID", paymentSession.UUID),
				logger.String("parentID", *paymentSession.ReferenceID),
			)
		}
		return
	}

	if paymentSession.IsAutoSplitPaymentFirstPayment() {
		if request.ChargeStatus == constant.ChargeStatusFailed {
			err = s.internalUnifiedPaymentSvc.AbortSplitPaymentOnCITFailure(ctx, &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: *paymentSession.ReferenceID,
				MerchantID:      paymentSession.MerchantID,
				MethodDetail:    request.ChargePaymentMethodDetails,
			})
			if err != nil {
				s.logger.Error(ctx, "failed to abort split payment", logger.Error(err), logger.String("paymentID", paymentSession.UUID))
			}
			return
		}

		if request.ChargeStatus != constant.ChargeStatusSuccess {
			return
		}

		err := s.internalUnifiedPaymentSvc.ContinueSplitPaymentExecution(ctx, &paymentModel.ProcessSplitPaymentRequest{
			ParentPaymentID: *paymentSession.ReferenceID,
			MerchantID:      paymentSession.MerchantID,
		})
		if err != nil {
			s.logger.Error(ctx, "failed to continue remaining split payment",
				logger.Error(err),
				logger.String("paymentID", paymentSession.UUID),
				logger.String("parentID", *paymentSession.ReferenceID),
			)
		}
	}

	err = s.internalUnifiedPaymentSvc.EvaluateSplitPaymentOutcome(ctx, paymentSession)
	if err != nil {
		s.logger.Error(ctx, "failed to evaluate split payment",
			logger.Error(err),
			logger.String("paymentID", paymentSession.UUID),
			logger.String("parentID", *paymentSession.ReferenceID),
		)
	}
}
