package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) Capture(ctx context.Context, request *unifiedPaymentModel.CaptureRequest) (*unifiedPaymentModel.CaptureResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/Capture")
	defer span.End()

	// Get payment session by id
	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	// Get charge by payment ID
	accountTrx, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	} else if accountTrx == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
	}

	// Validate process
	charge := unifiedPaymentModel.AccountTransactionToChargeResponse(accountTrx)
	if errValidate := s.validatePaymentForCapture(ctx, request, payment, charge); errValidate != nil {
		s.logger.Warn(ctx, "error when validate payment for capture", logger.Error(errValidate), logger.String("paymentId", request.PaymentID))
		return nil, errValidate
	}

	paymentCaptureID, _ := uuid.NewV7()
	paymentCapture := &paymentCaptureModel.PaymentCapture{
		ID:                     paymentCaptureID.String(),
		PaymentID:              payment.UUID,
		Status:                 constant.StatusPending,
		ReleaseRemainingAmount: request.ReleaseRemainingAmount,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}
	if request.Amount != nil {
		paymentCapture.Currency = request.Amount.Currency
		paymentCapture.Amount = request.Amount.Value
	}
	if err := s.paymentCaptureRepo.Insert(ctx, paymentCapture); err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	// Publish to async process
	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.PaymentCaptureProcessRoutingKey, nil, &unifiedPaymentModel.ProcessCaptureRequest{
		ID: paymentCapture.ID,
	})

	return &unifiedPaymentModel.CaptureResponse{
		ID:                              paymentCapture.ID,
		PaymentSessionID:                payment.UUID,
		PaymentSessionClientReferenceId: util.ValueOfPtr(payment.ReferenceID),
		ReleaseRemainingAmount:          request.ReleaseRemainingAmount,
		Amount:                          request.Amount,
		Status:                          constant.StatusPending,
		CreatedAt:                       paymentCapture.CreatedAt,
		UpdatedAt:                       paymentCapture.UpdatedAt,
	}, nil
}

func (s *UnifiedPaymentService) ProcessCapture(ctx context.Context, request *unifiedPaymentModel.ProcessCaptureRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ProcessCapture")
	defer span.End()

	s.logger.Info(ctx, "received process capture request", logger.Any("request", request))
	defer func() {
		s.logger.Info(ctx, "finish process capture request", logger.Any("request", request))
	}()

	// Implement exclusive lock for process capture by paymentCapture.ID
	queueKey := fmt.Sprintf("backend-portal:payment-capture:process:%s", request.ID)
	if ok, errLock := s.redis.SetNX(ctx, queueKey, true, 5*time.Minute).Result(); errLock != nil {
		s.logger.Error(ctx, "set exclusive queue with key "+queueKey, logger.Error(errLock))
		return pkgErr.New(response.HttpErrDatabase, errLock)

	} else if !ok {
		return pkgErr.New(response.HttpErrDupCheck, constant.ErrCaptureIsBeingProcessed)
	}

	// Get payment capture record
	paymentCapture, err := s.paymentCaptureRepo.GetByID(ctx, request.ID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if paymentCapture == nil {
		s.logger.Warn(ctx, "payment capture not found", logger.String("paymentCaptureId", request.ID))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentCaptureNotFound)
	}

	// Set distributed lock for processing payment capture for the same payment ID
	// This prevents concurrent captures on the same payment
	mutex := s.redis.NewMutex(
		"backend-portal:payment-capture:"+paymentCapture.PaymentID+":process:lock",
		redsync.WithExpiry(30*time.Second),
		redsync.WithRetryDelay(500*time.Millisecond),
		redsync.WithTries(60),
	)
	if err := mutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "unable to acquire payment lock, another capture may be in progress",
			logger.Error(err),
			logger.String("paymentId", paymentCapture.PaymentID),
			logger.String("captureId", request.ID))
		return pkgErr.New(response.HttpErrDupCheck, constant.ErrAnotherCaptureInProgress)
	}
	defer func() {
		if _, unlockErr := mutex.UnlockContext(ctx); unlockErr != nil {
			s.logger.Warn(ctx, "Failed to unlock payment process mutex", logger.Error(unlockErr))
		}
	}()

	// Get payment session by id
	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentCapture.PaymentID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	// Get charge by payment ID
	accountTrx, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if accountTrx == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
	}
	charge := unifiedPaymentModel.AccountTransactionToChargeResponse(accountTrx)
	charge.SetCaptureHistories(payment.PaymentCaptures)

	// Validate capture amount
	requestAmount := paymentCapture.Amount
	authorizedAmount := payment.Amount.InexactFloat64()
	capturedAmount := charge.Amount.Value
	if charge.CapturedAmount != nil {
		capturedAmount = charge.CapturedAmount.Value
	}
	if capturedAmount+requestAmount > authorizedAmount {
		return pkgErr.New(response.HttpErrRequest, constant.ErrCaptureAmountExceedAuthorizedAmount)
	}

	// Hit capture to processor
	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        serviceName,
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
	})
	processorCaptureResponse, err := s.creditCardProcessorRepo.Capture(ctx, &creditcardCoreProcessorModel.CaptureRequest{
		MerchantID:             payment.MerchantID,
		ClientTransactionID:    util.ValueOfPtr(payment.ReferenceID),
		AcquirerTransactionID:  accountTrx.ProcessorReferenceId,
		ReleaseRemainingAmount: paymentCapture.ReleaseRemainingAmount,
		Currency:               paymentCapture.Currency,
		Amount:                 paymentCapture.Amount,
		CapturedAmount:         capturedAmount,
	})
	if err != nil {
		return err
	}

	if processorCaptureResponse.Status != constant.StatusSuccess {
		s.logger.Warn(ctx, "processor capture response status is not SUCCESS", logger.String("paymentCaptureId", request.ID))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrProcessorCaptureStatusNotSuccess)
	}

	// Begin transaction for capture process
	ctxTx, err := s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := s.paymentRepo.RollbackTransaction(ctxTx); rollbackErr != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(rollbackErr))
			}
		}
	}()

	// Update account_transactions with captured amount info in metadata
	newCapturedAmount := capturedAmount + requestAmount

	// Update account transaction metadata with captured amount info
	updateRequest := orchestratorModel.UpdatePaymentTransactionRequest{
		LedgerId:             accountTrx.UUID.String(),
		ChargeStatus:         constant.ChargeStatusWaitingForCapture,
		UpdatedAt:            time.Now().UTC(),
		TransactionTimestamp: time.Now().UTC(),
	}

	// Update account transaction credit with new captured amount
	if err = s.accountTransactionRepo.UpdateCreditDebitByID(ctxTx, accountTrx.UUID.String(), &newCapturedAmount, nil); err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	// Set payment status to SUCCESS
	if newCapturedAmount == authorizedAmount {
		updateRequest.Status = constant.StatusSuccess
		updateRequest.ChargeStatus = constant.ChargeStatusSuccess
		payment.Status = constant.UnifiedPaymentSessionStatusPaid
	}

	// Update payment capture status to SUCCESS
	paymentCapture.Status = constant.StatusSuccess
	paymentCapture.UpdatedAt = time.Now().UTC()
	paymentCapture.ProcessorReferenceID = util.ValueToPtr(processorCaptureResponse.ID)
	if err = s.paymentCaptureRepo.Update(ctxTx, paymentCapture); err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	// Handle release remaining amount
	if paymentCapture.ReleaseRemainingAmount {
		if newCapturedAmount > 0 {
			// Update to SUCCESS status and payment to PAID
			updateRequest.Status = constant.StatusSuccess
			updateRequest.ChargeStatus = constant.ChargeStatusSuccess
			payment.Status = constant.UnifiedPaymentSessionStatusPaid
		} else {
			// Update to FAILED status and payment to CANCELED
			updateRequest.Status = constant.StatusFailed
			updateRequest.ChargeStatus = constant.ChargeStatusFailed
			payment.Status = constant.UnifiedPaymentSessionStatusCancelled

			if util.ValueOfPtr(payment.ExpiredAt).Before(time.Now().UTC()) {
				updateRequest.ChargeStatus = constant.ChargeStatusExpired
				payment.Status = constant.UnifiedPaymentSessionStatusExpired
			}
		}
	}

	if slices.Contains([]string{
		constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedPaymentSessionStatusCancelled,
		constant.UnifiedPaymentSessionStatusExpired}, payment.Status) {

		if err = s.paymentRepo.UpdatePaymentData(ctxTx, payment.ToDTO()); err != nil {
			return pkgErr.New(response.HttpErrDatabase, err)
		}

		// Modification of special payment channels for payments via credit cards
		if payment.PaymentMethod.Type == constant.ChannelCreditCard {
			merchant, err := s.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
			if err != nil {
				s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
				return err
			}

			binInformations := unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{}
			if charge.Card != nil {
				binInformations = charge.Card.BinInformations
			}

			channel := "FOREIGN_"
			if merchant.BusinessCountry.String == binInformations.Country {
				channel = "LOCAL_"
			}
			channel += strings.ToUpper(binInformations.Brand)

			payment.PaymentMethod.Acquirer = channel // For calculation payment fee per channel
		}

		if err := s.paymentSvc.DeterminePaymentFee(&ctxTx, payment); err != nil {
			return pkgErr.New(response.HttpErrDatabase, err)
		}

		updateRequest.Amount = commonModel.Amount{
			Currency: payment.Currency,
			Value:    fmt.Sprintf("%.2f", newCapturedAmount),
		}

		if err := s.paymentSvc.UpdatePendingLedger(ctxTx, payment, updateRequest); err != nil {
			s.logger.Error(ctx, "Failed to update pending ledger", logger.Error(err))
			return err
		}
	}

	// Commit transaction
	if err = s.paymentRepo.CommitTransaction(ctxTx); err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	// refresh payment data
	payment, err = s.paymentRepo.GetPaymentById(ctx, paymentCapture.PaymentID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	// Send capture as charge callback
	s.sendPaymentChargeCallback(ctx, charge.ID, payment)

	// Send payment status changed callback
	if slices.Contains([]string{
		constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedPaymentSessionStatusCancelled,
		constant.UnifiedPaymentSessionStatusExpired}, payment.Status) {

		s.SendCallback(ctx, payment)
	}

	return nil
}

func (s *UnifiedPaymentService) validatePaymentForCapture(
	ctx context.Context,
	request *unifiedPaymentModel.CaptureRequest,
	payment *paymentModel.Payment,
	charge *unifiedPaymentModel.ChargeResponse,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/validatePaymentForCapture")
	defer segment.End()

	// Validation merchant matching
	if payment.MerchantID != request.MerchantID {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	}

	// Validate payment method
	if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodTypeIsNotCard)
	}

	// Validate amount
	if request.Amount != nil && request.Amount.Currency == constant.CurrencyIDR && math.Mod(request.Amount.Value, 1) != 0 {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrAmountNotPermittedToUseDecimal)
	}

	// Parse payment metadata
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	// Validate capture method must be MANUAL
	if unifiedPaymentMetadata.PaymentMethodOptions.Card == nil ||
		strings.ToUpper(unifiedPaymentMetadata.PaymentMethodOptions.Card.CaptureMethod) != constant.UnifiedPaymentCardCaptureMethodManual {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrCaptureMethodMustBeManual)
	}

	// validate payment status
	if slices.Contains([]string{
		constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedPaymentSessionStatusCancelled,
		constant.UnifiedPaymentSessionStatusExpired,
	}, payment.Status) {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentAlreadyInFinalStatus)
	}

	// Validate charge status
	if charge.Status != constant.ChargeStatusWaitingForCapture {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrChargeStatusMustBeWaitingForCapture)
	}

	// Validate capture amount
	if request.Amount != nil {
		if request.Amount.Currency != payment.Currency {
			return pkgErr.New(response.HttpErrRequest, constant.ErrCurrencyIsNotMatch)
		}

		requestAmount := request.Amount.Value
		authorizedAmount := payment.Amount.InexactFloat64()
		capturedAmount := charge.Amount.Value
		if charge.CapturedAmount != nil {
			capturedAmount = charge.CapturedAmount.Value
		}

		if capturedAmount+requestAmount > authorizedAmount {
			return pkgErr.New(response.HttpErrRequest, constant.ErrCaptureAmountExceedAuthorizedAmount)
		}
	}

	return nil
}
