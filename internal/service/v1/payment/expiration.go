package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreQrModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"

	constant "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// PublishPaymentExpirationMessage will publish a message to RabbitMQ to mark payment as expired using DLQ
// This function will be called by a cron job only and will be executed once a day
// payment that will be publish is payment that will be expired in the next 24 hours (01:00 - 00:59:59 the next day)
func (s *PaymentService) PublishPaymentExpirationMessage(ctx context.Context) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/PublishPaymentExpirationMessage")
	defer segment.End()

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)

	payments, err := s.paymentRepo.GetExpiringPayments(ctx, start, end)
	if err != nil {
		return err
	}

	if payments == nil {
		return nil
	}

	for _, expiringPayment := range payments {
		duration := expiringPayment.ExpiredAt.Sub(time.Now().UTC())
		err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			expiringPayment,
			duration,
		)

		if err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
			return err
		}
	}

	return nil
}

// ExpirePayment updates the status of a payment to expired if it meets certain conditions.
// It first retrieves the payment by UUID and performs several validations:
// - Checks if the payment exists
// - Verifies the merchant ID matches the payment's merchant ID
// - Ensures the payment is not already in a final status (success, void, expired, or failed)
// If all validations pass, it updates the payment status to expired.
// It also updates the chargeStatus in account_transaction for payment v2 compatibility.
func (s *PaymentService) ExpirePayment(ctx context.Context, request paymentModel.ExpiringPayment) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ExpirePayment")
	defer segment.End()

	s.logger.Info(ctx, "process expiring payment", logger.String("paymentID", request.UUID))

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.UUID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Warn(ctx, "payment not found", logger.Any("request", request))
		return nil
	}

	if payment.MerchantID != request.MerchantID {
		s.logger.Warn(ctx, "merchant is not match", logger.Any("request", request))
		return nil
	}

	if payment.Status == constant.UnifiedPaymentSessionStatusProcessing {
		if metadata := payment.ToUnifiedPaymentMetadata(); metadata != nil && metadata.IsAutoSplitPaymentAuth() {
			s.logger.Info(ctx, "Payment expiration handling skipped due to split payment authentication transaction", logger.Any("request", request))
			return nil
		}
		if request.ChargeStatus == constant.ChargeStatusProcessing {
			return s.handleProcessedPayment(ctx, payment)
		} else if request.ChargeStatus == constant.ChargeStatusWaitingForCapture {
			return s.handleAuthorizedPayment(ctx, payment)
		}
	}

	if payment.PaymentMethod.Type == constant.UnifiedPaymentMethodEWallet {
		paymentAfterInquiry, err := s.unifiedPaymentSvc.InquiryEWalletPayment(ctx, payment)
		if err != nil {
			s.logger.Error(ctx, "error when inquiry ewallet payment status", logger.Error(err), logger.String("paymentID", payment.UUID))
			return err
		}

		if paymentAfterInquiry != nil && paymentAfterInquiry.InquiryDetail != nil && paymentAfterInquiry.InquiryDetail.Status == constant.StatusFailed {
			request.ChargeStatus = constant.ChargeStatusExpired
			// Continue to expired payment processing below
		} else {
			// Only handle non-expired EWallet payments here
			if payment.Status == constant.UnifiedPaymentSessionStatusRequireAction &&
				payment.ToUnifiedPaymentResponse().Mode == constant.UnifiedPaymentModeAPI {
				_, err = s.unifiedPaymentSvc.UpdateEWalletPaymentSession(ctx, payment.UUID)
				if err != nil {
					s.logger.Error(ctx, "error when update ewallet payment session to processing for API mode payment", logger.Error(err), logger.String("paymentID", payment.UUID))
					return err
				}
			}
			return nil
		}
	}

	if slices.Contains([]string{
		paymentConstant.PAYMENT_STATUS_SUCCESS,
		paymentConstant.PAYMENT_STATUS_VOID,
		paymentConstant.QrisStatusExpired,
		paymentConstant.QrisStatusFailed,
		constant.UnifiedPaymentSessionStatusExpired,
		constant.UnifiedPaymentSessionStatusCancelled,
		constant.UnifiedPaymentSessionStatusPaid,
	}, payment.Status) {
		s.logger.Warn(ctx, "payment already in final status", logger.Any("request", request))
		return nil
	}

	// For VA sync call to snap-core to check if already paid before expiring
	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		vaNumber := getVANumberFromPayment(payment)
		if vaNumber == "" {
			// NULL, then continue with expiry
		} else {
			resp, err := s.snapCoreRepo.InquiryStatusVirtualAccount(ctx, &snapCoreVAModel.InquiryStatusVARequest{
				VirtualAccount: vaNumber,
				SkipPublish:    false,
			})

			if err != nil {
				// ERROR on the inquiry, skipping expiry
				// to avoid double callback
				s.logger.Error(ctx,
					"error when inquiry VA payment status",
					logger.Error(err),
					logger.String("paymentID", payment.UUID))
				return nil
			}

			if resp != nil {
				if resp.IsNotFound() {
					s.logger.Warn(ctx, "VA not found, go with expiry", logger.String("paymentID", payment.UUID), logger.String("vaNumber", vaNumber))
					err = nil
				}
				if resp.IsConflict() {
					s.logger.Info(ctx, "VA already paid in processor, skipping expiration", logger.String("paymentID", payment.UUID), logger.String("vaNumber", vaNumber))
					return nil
				}
				if resp.IsPaid() {
					s.logger.Info(ctx, "VA payment already paid, skipping expiration", logger.String("paymentID", payment.UUID))
					return nil
				}
			}
		}
	}

	// For QRIS sync call to snap-core to check if already final status before expiring
	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_QRIS {
		charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
		if err != nil || charge == nil || charge.ProcessorReferenceId == "" {
			// continue with expiry
			s.logger.Warn(ctx, "Failed to find existing charge for payment",
				logger.String("paymentUUID", payment.UUID),
				logger.Error(err),
			)
		} else {
			qrisID := charge.ProcessorReferenceId
			resp, err := s.snapCoreRepo.InquiryStatusQris(ctx, &snapCoreQrModel.InquiryStatusQrMpmRequest{
				QrisUUID:    qrisID,
				SkipPublish: false,
			})

			if err != nil || resp == nil {
				// ERROR on the inquiry, skipping expiry
				// to avoid double callback
				s.logger.Warn(
					ctx,
					"Error in QR inquiry status",
					logger.String("qrisID", qrisID),
					logger.Error(err))
				return nil
			}

			latestStatus := util.MapQRLatestStatusToPaymentStatus(resp.Data.Status)

			// Final Status, Skip Expiration
			if latestStatus == constant.StatusSuccess || latestStatus == constant.StatusFailed {
				s.logger.Info(
					ctx,
					"QRIS already final status in processor, skipping expiry",
					logger.String("paymentID", payment.UUID),
					logger.String("qrisID", qrisID),
					logger.String("latestStatus", latestStatus))
				return nil
			}
		}
	}

	if err = s.processExpiration(ctx, payment, request.ChargeStatus); err != nil {
		return err
	}

	s.unifiedPaymentSvc.SendCallback(ctx, payment)

	return nil
}

func (s *PaymentService) handleProcessedPayment(ctx context.Context, payment *paymentModel.Payment) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/handleProcessedPayment")
	defer segment.End()

	var (
		ccPaymentNotifRequest = creditcardModel.CardPaymentNotificationRequest{
			Event: "UPDATE_PAYMENT_STATUS",
		}
		merchantID = payment.MerchantID
	)

	// this handle currently exclusive for CC
	// other than it, should return immediately
	if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		s.logger.Warn(ctx, "payment method did not have procesing expiration handler", logger.String("method", payment.PaymentMethod.Name), logger.String("paymentID", payment.UUID))
		return nil
	}

	charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.ReferencePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment charge", logger.Error(err))
		return nil
	}

	if charge == nil {
		s.logger.Error(ctx, "payment charge not found", logger.String("paymentID", payment.UUID))
		return nil
	}

	if parentID := payment.GetOnBehalfParentID(); parentID != "" {
		merchantID = parentID
	}

	response, err := s.creditCardSvc.InquiryTransaction(ctx, &creditcardModel.InquiryTransactionRequest{
		MerchantID:           merchantID,
		ClientReferenceID:    util.ValueOfPtr(payment.ReferenceID),
		ProcessorReferenceID: charge.ProcessorReferenceId,
	})

	if err != nil {
		s.logger.Error(ctx, "inquiry result return error and force payment to failed")
		trxID, err := uuid.Parse(charge.ProcessorTransactionId)
		if err != nil {
			s.logger.Error(ctx, "failed to parse processor trx id", logger.Error(err), logger.String("trxID", charge.ProcessorTransactionId))
		}
		response = &creditcardModel.PaymentNotificationDataRequest{
			PaymentUUID:           uuid.MustParse(payment.UUID),
			ReferenceID:           *payment.ReferenceID,
			PaymentStatus:         constant.UnifiedPaymentSessionStatusCancelled,
			Amount:                payment.Amount,
			AcquirerTransactionID: charge.ProcessorReferenceId,
			TransactionID:         trxID,
			Updated:               time.Now().UTC(),
		}
	}

	// when the payment still in processing status, we need to republish the expiration message
	// with the delayed config based on the retry count
	// so the payment will be re-checked again in the future
	// if the payment still in processing status, it will be republished again until the delayed config is zero
	// which means we should not republish it again
	if response.PaymentStatus == constant.UnifiedPaymentSessionStatusProcessing {
		s.logger.Info(ctx, "payment still in processing after inquiry", logger.String("paymentID", payment.UUID))

		retryCount, delayDuration := s.getDelayedConfigDuration(ctx, payment)
		if delayDuration == 0 {
			s.logger.Info(ctx, "payment method is not have delayed config for processing payment, skip re-publish expiration message")
			return nil
		}

		if delayDuration == -1 {
			s.logger.Info(ctx, "the delay duration backoff iteration was exceeded, cancel the payment", logger.String("paymentID", payment.UUID), logger.Int("retryCount", retryCount))

			trxID, _ := uuid.Parse(charge.ProcessorTransactionId)
			ccPaymentNotifRequest.Data = creditcardModel.PaymentNotificationDataRequest{
				PaymentUUID:           uuid.MustParse(payment.UUID),
				ReferenceID:           util.ValueOfPtr(payment.ReferenceID),
				PaymentStatus:         constant.UnifiedPaymentSessionStatusCancelled,
				Amount:                payment.Amount,
				AcquirerTransactionID: charge.ProcessorReferenceId,
				TransactionID:         trxID,
				Updated:               time.Now().UTC(),

				// seeding data from inquiry result to make it up to date
				// TODO: move the expiration logic into processor
				AuthenticationMethod: response.AuthenticationMethod,
				AuthenticationData:   response.AuthenticationData,
				AuthorizationData:    response.AuthorizationData,
				ResponseCode:         response.ResponseCode,
				BankMerchantID:       response.BankMerchantID,
				MerchantID:           response.MerchantID,
				CardData:             response.CardData,
				MIDInfo:              response.MIDInfo,
				Type:                 response.Type,
				Currency:             response.Currency,
				PaymentURL:           response.PaymentURL,
				RedirectUrl:          response.RedirectUrl,
			}
			bytes, err := json.Marshal(ccPaymentNotifRequest)
			if err != nil {
				s.logger.Error(ctx, "failed to marshal cc inquiry response", logger.Error(err))
				return nil
			}

			err = s.rabbitMqExt.Publish(ctx, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, nil, bytes)
			if err != nil {
				s.logger.Error(ctx, "failed to publish payment notification", logger.Error(err))
				return nil
			}
		}

		err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:         payment.UUID,
				MerchantID:   payment.MerchantID,
				Note:         fmt.Sprintf("Payment already processed at %v (UTC) with retry count %v", time.Now(), retryCount),
				ChargeStatus: constant.ChargeStatusProcessing,
			},
			delayDuration,
		)
		if err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
		}
		return nil
	}

	ccPaymentNotifRequest.Data = *response

	bytes, err := json.Marshal(ccPaymentNotifRequest)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal cc inquiry response", logger.Error(err))
		return nil
	}

	err = s.rabbitMqExt.Publish(ctx, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, nil, bytes)
	if err != nil {
		s.logger.Error(ctx, "failed to publish payment notification", logger.Error(err))
		return nil
	}

	return nil
}

func (s *PaymentService) handleAuthorizedPayment(ctx context.Context, payment *paymentModel.Payment) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/handleAuthorizedPayment")
	defer segment.End()

	// this handle currently exclusive for CC
	// other than it, should return immediately
	if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		s.logger.Warn(ctx, "payment method did not have processing expiration handler", logger.String("method", payment.PaymentMethod.Name), logger.String("paymentID", payment.UUID))
		return nil
	}

	charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.ReferencePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment charge", logger.Error(err))
		return nil
	}

	if charge == nil {
		s.logger.Error(ctx, "payment charge not found", logger.String("paymentID", payment.UUID))
		return nil
	}

	chargeResponse := unifiedPaymentModel.AccountTransactionToChargeResponse(charge)
	if chargeResponse.Status != constant.ChargeStatusWaitingForCapture {
		s.logger.Warn(ctx, "charge status not in waiting for capture", logger.String("paymentID", payment.UUID))
		return nil
	}

	// If expiration is max expiry handling for card, then only cancel/expire without voiding the transactions.
	if time.Now().UTC().Before(util.ValueOfPtr(payment.ExpiredAt)) {
		inquiryResponse, err := s.creditCardSvc.InquiryTransaction(ctx, &creditcardModel.InquiryTransactionRequest{
			MerchantID:           payment.MerchantID,
			ClientReferenceID:    util.ValueOfPtr(payment.ReferenceID),
			ProcessorReferenceID: charge.ProcessorReferenceId,
		})

		if err != nil {
			s.logger.Error(ctx, "inquiry result return error and force payment to failed", logger.String("paymentID", payment.UUID))
			trxID, err := uuid.Parse(charge.ProcessorTransactionId)
			if err != nil {
				s.logger.Error(ctx, "failed to parse processor trx id", logger.Error(err), logger.String("trxID", charge.ProcessorTransactionId))
			}
			inquiryResponse = &creditcardModel.PaymentNotificationDataRequest{
				PaymentUUID:           uuid.MustParse(payment.UUID),
				ReferenceID:           *payment.ReferenceID,
				PaymentStatus:         constant.UnifiedPaymentSessionStatusCancelled,
				Amount:                payment.Amount,
				AcquirerTransactionID: charge.ProcessorReferenceId,
				TransactionID:         trxID,
				Updated:               time.Now().UTC(),
			}
		}

		if inquiryResponse == nil {
			s.logger.Error(ctx, "inquiry response is empty", logger.String("paymentID", payment.UUID))
			return constant.ErrDataNotFound
		}

		if inquiryResponse.PaymentStatus == constant.StatusSuccess {
			s.logger.Info(ctx, "inquiry response has already succeeded", logger.String("paymentID", payment.UUID))
			return nil
		}

		if inquiryResponse.PaymentStatus == constant.UnifiedPaymentSessionStatusProcessing &&
			slices.Contains([]string{constant.CardTransactionStatusAuthorized, constant.CardTransactionStatusPartiallyCaptured}, util.ValueOfPtr(inquiryResponse.AuthorizationData).TransactionStaus) {

			retryExpiryInMinutes := s.config.UnifiedPaymentConfig.RetryExpiringAuthorizedTransactionMinutes
			delayDuration := time.Duration(retryExpiryInMinutes) * time.Minute
			s.logger.Info(ctx, fmt.Sprintf("inquiry transaction status still in PENDING, try to process expiry transaction in %d minutes", retryExpiryInMinutes), logger.String("paymentID", payment.UUID))
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
				s.logger.Error(ctx, "error publish payment expiration in waiting for charge status message", logger.Error(err))
			}

			return nil
		}

		// Begin transaction for capture process
		ctxTx, err := s.paymentRepo.BeginTransaction(ctx)
		if err != nil {
			return pkgErrors.New(response.HttpErrDatabase, err)
		}
		defer func() {
			if err != nil {
				if rollbackErr := s.paymentRepo.RollbackTransaction(ctx); rollbackErr != nil {
					s.logger.Error(ctx, "failed to rollback transaction", logger.Error(rollbackErr))
				}
			}
		}()

		// Force update payment status = PAID and charge status SUCCESS if captured amount > 0
		if util.ValueOfPtr(chargeResponse.CapturedAmount).Value > 0 {
			s.logger.Info(ctx, "force release for authorized transaction without voiding transaction", logger.String("paymentID", payment.UUID))

			payment.Status = constant.UnifiedPaymentSessionStatusPaid
			if err = s.paymentRepo.UpdatePaymentData(ctxTx, payment.ToDTO()); err != nil {
				return pkgErrors.New(response.HttpErrDatabase, err)
			}

			if err := s.DeterminePaymentFee(&ctxTx, payment); err != nil {
				return pkgErrors.New(response.HttpErrDatabase, err)
			}

			updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
				LedgerId:             charge.UUID.String(),
				ChargeStatus:         constant.ChargeStatusSuccess,
				UpdatedAt:            time.Now().UTC(),
				TransactionTimestamp: time.Now().UTC(),
				Amount: commonModel.Amount{
					Currency: payment.Currency,
					Value:    fmt.Sprintf("%.2f", util.ValueOfPtr(chargeResponse.CapturedAmount).Value),
				},
			}

			if err := s.UpdatePendingLedger(ctxTx, payment, updateRequest); err != nil {
				s.logger.Error(ctx, "Failed to update pending ledger", logger.Error(err))
				return err
			}

			// Commit transaction
			if err = s.accountTransactionRepo.CommitTransaction(ctxTx); err != nil {
				return pkgErrors.New(response.HttpErrDatabase, err)
			}

			s.unifiedPaymentSvc.SendCallback(ctx, payment)

			return nil
		}

		// Expire payment and charge.
		if err = s.processExpiration(ctxTx, payment, constant.ChargeStatusExpired); err != nil {
			return err
		}

		// Commit transaction
		if err = s.paymentRepo.CommitTransaction(ctxTx); err != nil {
			return pkgErrors.New(response.HttpErrDatabase, err)
		}

		// Send callback
		s.unifiedPaymentSvc.SendCallback(ctx, payment)

		return nil
	}

	s.logger.Info(ctx, "processing payment expiration for authorized transaction", logger.String("paymentID", payment.UUID))
	if _, err = s.unifiedPaymentSvc.Capture(ctx, &unifiedPaymentModel.CaptureRequest{
		PaymentID:              payment.UUID,
		ChargeID:               chargeResponse.ID,
		MerchantID:             payment.MerchantID,
		ReleaseRemainingAmount: true,
	}); err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) processExpiration(ctx context.Context, payment *paymentModel.Payment, chargeStatus string) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/processExpiration")
	defer segment.End()

	expiredStatus := paymentConstant.PaymentStatusExpired
	if payment.Status == constant.UnifiedStaticPaymentStatusActive {
		expiredStatus = constant.UnifiedStaticPaymentStatusInactive
	}

	payment.Status = expiredStatus
	if err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, time.Now().UTC()); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	s.recordPaymentExpired(ctx, payment.UUID, constant.StatusHistoryActorSystem)

	// Skip for static payment
	if expiredStatus == constant.UnifiedStaticPaymentStatusInactive {
		return nil
	}

	if chargeStatus == "" {
		chargeStatus = constant.ChargeStatusExpired
	}

	paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "error when find payment ledger by reference", logger.Error(err), logger.String("paymentID", payment.UUID))
		return pkgErrors.New(response.HttpStatusErrorNotFound, err)
	} else if paymentLedger != nil {
		metadata := orchestrator_model.MetadataPayment[any]{
			ChargeStatus: chargeStatus,
		}

		err = s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(
			ctx,
			orchestrator_model.UpdatePaymentTransactionRequest{
				LedgerId:    paymentLedger.UUID.String(),
				Status:      constant.StatusFailed,
				FailureCode: constant.FailureCodeChargeExpired,
			},
			metadata,
		)
		if err != nil {
			s.logger.Error(ctx, "error updating account transaction chargeStatus", logger.Error(err))
		}
	}

	subject, message := constant.GetNotificationMessage(payment.UUID, expiredStatus)
	err = s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, payment.UUID),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   subject,
			Type:      constant.CreateVAPaymentNotifType,
			Message:   message,
			CreatedAt: time.Now().UTC(),
			Status:    expiredStatus,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "push notification for payment "+payment.UUID, logger.Error(err))
	}

	return nil
}

func (s *PaymentService) HandleStrictExpiry(ctx context.Context, paymentID string) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/HandleStrictExpiry")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Warn(ctx, "payment not found", logger.Any("paymentID", paymentID))
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	if payment.ToUnifiedPaymentResponse().ExpirationMode != constant.UnifiedPaymentExpirationModeStrict {
		return pkgErrors.New(response.HttpErrRequest, errors.New("payment expiration mode is not strict"))
	}

	expiredStatus := paymentConstant.PaymentStatusExpired
	if payment.Status == constant.UnifiedStaticPaymentStatusActive {
		expiredStatus = constant.UnifiedStaticPaymentStatusInactive
	}

	if err = s.paymentRepo.UpdatePaymentStatus(ctx, paymentID, payment.MerchantID, expiredStatus, time.Now().UTC()); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	s.recordPaymentExpired(ctx, paymentID, constant.StatusHistoryActorSystem)

	chargeStatus := constant.ChargeStatusExpired

	paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, paymentID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "error when find payment ledger by reference", logger.Error(err), logger.String("paymentID", paymentID))
		return pkgErrors.New(response.HttpStatusErrorNotFound, err)
	} else if paymentLedger != nil {
		metadata := orchestrator_model.MetadataPayment[any]{
			ChargeStatus: chargeStatus,
		}

		err = s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(
			ctx,
			orchestrator_model.UpdatePaymentTransactionRequest{
				LedgerId: paymentLedger.UUID.String(),
				Status:   constant.StatusFailed,
			},
			metadata,
		)
		if err != nil {
			s.logger.Error(ctx, "error updating account transaction chargeStatus", logger.Error(err))
			return pkgErrors.New(response.HttpErrDatabase, err)
		}
	}

	return nil
}

func getVANumberFromPayment(payment *paymentModel.Payment) string {
	if payment.ProcessorReferenceNumber != nil && *payment.ProcessorReferenceNumber != "" {
		return *payment.ProcessorReferenceNumber
	}

	if payment.Metadata != nil {
		if snapCore, ok := (*payment.Metadata)["snapCore"].(map[string]any); ok {
			if number, ok := snapCore["number"].(string); ok {
				return number
			}
		}
	}
	return ""
}
