package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) VCCTerminalBatchCharge(ctx context.Context, request *model.VCCTerminalChargeRequest) (*model.VCCTerminalBatchChargeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/VCCTerminalBatchCharge")
	defer segment.End()

	paymentMethodRequest := model.GetPaymentMethodFilterRequest{
		MerchantID: request.MerchantID,
		Category:   constantPayment.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       constantPayment.PAYMENT_METHOD_VIRTUAL_TERMINAL,
	}
	paymentMethod, err := s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, &paymentMethodRequest)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if paymentMethod == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrFeatureIsNotYetEnable)
	}

	bookings := []model.BookingPayload{}
	if err := s.internal.DecryptRequest(ctx, request.EncryptedRequest, &bookings); err != nil {
		return nil, err // The error message is already wrapped inside the function.
	}
	defer func() { bookings = nil }()

	partnerConfigs := paymentMethod.GetCardPartnerConfigForOnlineTravelAgent(s.config.VccTerminal.DefaultConfig)

	if err := s.validateVCCTerminalBatchCharge(partnerConfigs, bookings); err != nil {
		return nil, err // The error message is already wrapped inside the function.
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "VCC-Terminal",
		ReferenceId: request.MerchantID,
	})

	certPEM, err := s.creditCardSvc.GetCardEncryptionPublicKey(ctx, request.MerchantID)
	if err != nil {
		return nil, err // The error message is already wrapped inside the function.
	}

	var (
		batchID = util.GenerateUUID()
		wg, mx  = new(sync.WaitGroup), new(sync.Mutex)
		result  = &model.VCCTerminalBatchChargeResponse{BatchID: batchID.String()}
	)

	recordFailedProcess := func(booking model.BookingPayload) {
		mx.Lock()
		result.FailedCount++
		result.FailedTotal += booking.Amount.Value.InexactFloat64()
		result.FailedCharges = append(result.FailedCharges, booking.ToResponse())
		mx.Unlock()
	}

	workers, err := ants.NewPoolWithFuncGeneric(s.config.VccTerminal.WorkerCount, func(booking model.BookingPayload) {

		defer wg.Done()

		booking.MerchantID = request.MerchantID
		booking.BatchID = batchID.String()
		booking.PaymentID = util.GenerateUUID().String()
		booking.CreatedBy = request.UserID

		s.logger.Info(ctx, "Processing VCC terminal transaction", logger.Any("booking", booking.ToLog()))

		chargePayloadBytes, _ := json.Marshal(booking.ToCardChargePayload())

		encryptedPayload, err := s.cryptoProvider.EncryptPKCS7(certPEM, chargePayloadBytes)
		if err != nil {
			recordFailedProcess(booking)
			s.logger.Warn(ctx, "Failed to encrypt VCC terminal charge using PKCS#7", logger.Error(err))
			return
		}

		session, err := s.unifiedPaymentSvc.CreateSession(ctx, booking.ToCreateUnifiedPaymentSessionRequest(s.config.VccTerminal))
		if err != nil {
			recordFailedProcess(booking)
			s.logger.Warn(ctx, "Failed to create payment session for booking transaction", logger.Error(err))
			return
		}
		if len(session.ChargeDetails) > 0 {
			booking.ChargeID = session.ChargeDetails[0].ID
		}

		payload := booking.ToVCCTerminalChargeMessage(encryptedPayload)
		if err := s.rabbitMqExt.Publish(ctx, rabbitMqExt.VccTerminalChargeRoutingKey, nil, payload); err != nil {
			recordFailedProcess(booking)
			s.logger.Warn(ctx, "Failed to publish booking transaction", logger.Error(err))
			return
		}

		mx.Lock()
		result.SuccessCount++
		result.SuccessTotal += booking.Amount.Value.InexactFloat64()
		mx.Unlock()
	})
	if err != nil {
		s.logger.Error(ctx, "Failed to init worker pool for publish vcc-terminal transaction", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	defer workers.Release()

	for _, booking := range bookings {
		wg.Add(1)

		_ = workers.Invoke(booking)
	}

	wg.Wait()

	return result, nil
}

func (s *PaymentService) validateVCCTerminalBatchCharge(partnerConfigs map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj, bookings []model.BookingPayload) error {
	if len(bookings) == 0 {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("empty bookings transaction"))
	}

	for i, booking := range bookings {
		if err := s.validator.Struct(booking); err != nil {
			return pkgErrors.New(response.HttpErrRequest, err)
		}

		if _, found := s.config.VccTerminal.TravelAgents[booking.TravelAgentCode]; !found {
			return pkgErrors.New(response.HttpErrUnprocessableContent, fmt.Errorf("travel agent code %s not found", booking.TravelAgentCode))
		}

		config, ok := partnerConfigs[booking.TravelAgentCode]
		if !ok {
			config = partnerConfigs[constant.DefaultConfig]
		}
		isAllowedAllBinNumbers := len(config.SupportedUseCase.AllowedBinNumbers) == 1 &&
			config.SupportedUseCase.AllowedBinNumbers[0] == "ALL"
		if !isAllowedAllBinNumbers {
			allowedBinNumber := false
			for _, binNumber := range config.SupportedUseCase.AllowedBinNumbers {
				if binNumber == booking.Card.Number[:len(binNumber)] {
					allowedBinNumber = true
					break
				}
			}
			if !allowedBinNumber {
				return pkgErrors.New(response.HttpErrUnprocessableContent, fmt.Errorf(
					"card BIN is not allowed for travel agent code %s or global config, masked card number %s",
					booking.TravelAgentCode, booking.ToResponse().Card.Number),
				)
			}
		}
		bookings[i].AllowedCardTypes = config.CardTypes
		bookings[i].BankMerchantID = config.AcquirerMerchantID
		bookings[i].AllowedPrincipal = config.PrincipalAvailable
		bookings[i].AllowedBinNumbers = config.SupportedUseCase.AllowedBinNumbers
	}

	return nil
}

func (s *PaymentService) VCCTerminalSubmitCharge(ctx context.Context, request model.VCCTerminalChargeMessage) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/VCCTerminalSubmitCharge")
	defer segment.End()

	chargeKey := fmt.Sprintf(constant.VCCTerminalSubmitChargeCacheKey, request.PaymentID)
	if ok, err := s.redis.SetNX(ctx, chargeKey, true, constant.VCCTerminalSubmitChargeTTL).Result(); err != nil {
		return fmt.Errorf("set vcc terminal lock: %w", err)

	} else if !ok {
		return pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction has been processed"))
	}
	defer func() {
		if err == nil {
			return
		}
		if e := s.redis.Del(ctx, chargeKey).Err(); e != nil {
			s.logger.Error(ctx, "Failed to delete VCC terminal lock for payment ID "+request.PaymentID, logger.Error(e))
		}
	}()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		return fmt.Errorf("get payment by id: %w", err)

	} else if payment == nil {
		return pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction not found"))

	} else if payment.Status != constant.UnifiedPaymentSessionStatusProcessing {
		return pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction is already in final status"))
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "VCC-Terminal",
		ReferenceId: request.MerchantID,
		OriginId:    request.PaymentID,
	})

	authRequest := creditcardCoreProcessorModel.AuthenticationRequest{
		MerchantID:       request.MerchantID,
		PaymentID:        request.PaymentID,
		EncryptedPayload: request.EncryptedPayload,
	}
	if _, err = s.creditCardSvc.Authentication(ctx, authRequest); err != nil {
		// If the retry handling mechanism for messages has been implemented, payment status updates should only apply to non-retryable errors.
		func() {
			updateErr := s.paymentRepo.UpdatePaymentStatusWithReason(ctx, request.PaymentID, model.UpdatePaymentStatusWithReasonRequest{
				Status:            constant.UnifiedPaymentSessionStatusCancelled,
				ReasonType:        util.ValueToPtr(constant.CreditCardAuthorizationFailed),
				ReasonDescription: util.ValueToPtr("Failed to process authorization request"),
			})
			if updateErr != nil {
				s.logger.Warn(ctx, "Failed to update payment status with reason", logger.Error(updateErr))
				return
			}
			updateErr = s.accountTransactionRepo.UpdateStatusAccountTransaction(ctx, request.ChargeID, constant.StatusFailed, nil, nil)
			if updateErr != nil {
				s.logger.Warn(ctx, "Failed to update charge status", logger.Error(updateErr))
				return
			}
		}()
		s.recordPaymentCancelled(ctx, request.PaymentID, constant.StatusHistoryActorSystem)
		return err
	}
	return nil
}
