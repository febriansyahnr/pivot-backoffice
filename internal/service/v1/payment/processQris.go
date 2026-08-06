package paymentService

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
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) ProcessQrisPayment(ctx context.Context, request *paymentModel.QrisPaymentNotificationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ProcessQrisPayment")
	defer segment.End()

	s.logger.Info(ctx, "Process QRIS payment", logger.Any("request", request))

	// Get active payment by processorReferenceNumber
	payment, err := s.paymentRepo.GetActivePaymentByProcessorReferenceNumber(ctx, &paymentModel.GetActivePaymentByProcessorReferenceNumberRequest{
		ProcessorReferenceNumber: request.ReferenceNo,
	})
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// sometime the internal expiration process is faster than the notification
	// need to handle when the payment already expired but the notification still come
	// in this case, we just ignore the notification due get payment by processor ref number will always return nil when expired
	if payment == nil && strings.EqualFold(request.Status, paymentConstant.QrisStatusExpired) {
		return nil
	}

	if payment == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)
	}
	allowedPaymentStatus := []string{paymentConstant.QrisStatusPending, paymentConstant.QrisStatusExpired}
	if !util.Contains(allowedPaymentStatus, payment.Status) {
		s.logger.Info(ctx, "Payment is not pending or expired, skip processing", logger.Any("paymentId", payment.UUID), logger.String("status", payment.Status))
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentInFinalState)
	}

	if err = s.internal.DeterminePaymentFee(&ctx, payment); err != nil {
		return err
	}

	qrisMetadata, errQrisMetadata := buildQrisMetadata(payment.Metadata)
	if errQrisMetadata != nil {
		s.logger.Error(ctx, "Error build QRIS metadata", logger.Error(errQrisMetadata))
		return errQrisMetadata
	}

	if qrisMetadata.QrType == constant.QrTypeDynamic {
		lockKey := fmt.Sprintf(constant.PaymentNotificationLockCacheKey, payment.UUID)
		s.logger.Info(ctx, "acquire lock payment notification", logger.String("key", lockKey))
		mutex := s.redis.NewMutex(lockKey, redsync.WithExpiry(constant.PaymentNotificationLockTTL))
		err = mutex.LockContext(ctx)
		if err != nil {
			s.logger.Error(ctx, "error when acquire lock payment notification", logger.Error(err), logger.String("key", lockKey))
			return pkgErrors.New(response.HttpErrInternal, err)
		}
		defer func() {
			s.logger.Info(ctx, "release lock payment notification", logger.String("key", lockKey))
			_, errUnlock := mutex.UnlockContext(ctx)
			if errUnlock != nil {
				s.logger.Error(ctx, "error when release payment notification lock", logger.Error(errUnlock), logger.String("key", lockKey))
			}
		}()
	}

	// TODO: Handle CPM QRIS later, now skip for CPM
	if qrisMetadata.QrMethodType == constant.QrMethodTypeCPM {
		s.logger.Error(ctx, "Payment notification for QR CPM is not ready", logger.Any("paymentId", payment.UUID))
		return pkgErrors.New(response.HttpErrRequest, errors.New("QR method type is not ready"))
	}

	if !slices.Contains([]string{constant.QrTypeDynamic, constant.QrTypeStatic}, qrisMetadata.QrType) {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidQrType)
	}

	if !request.TransactionTime.IsZero() {
		payment.TrxDatetime = &request.TransactionTime
	}
	payment.ProcessorTransactionID = request.ProcessorTransactionID

	switch strings.ToUpper(request.Status) {
	case paymentConstant.QrisStatusSuccess:
		return s.payQrisPayment(ctx, request, payment, qrisMetadata)

	case paymentConstant.QrisStatusFailed:
		return s.failQrisPayment(ctx, request, payment, qrisMetadata)

	case paymentConstant.QrisStatusPending:
		return nil // Do noting on PENDING status

	case paymentConstant.QrisStatusExpired:
		return nil

	default:
		return pkgErrors.New(response.HttpErrRequest, constant.ErrStatusNotAllowed)
	}
}

func (s *PaymentService) payQrisPayment(ctx context.Context, request *paymentModel.QrisPaymentNotificationRequest, payment *paymentModel.Payment, qrisMetadata *paymentModel.PaymentMetadataQris) error {

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/payQrisPayment")
	defer segment.End()

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

	// Noting payment update for QR Static
	if qrisMetadata.QrType == constant.QrTypeDynamic {
		if errPayQrMpmDynamic := s.payQrisMpmDynamic(ctxTrx, request, payment); errPayQrMpmDynamic != nil {
			return errPayQrMpmDynamic
		}
	}
	payment.Processor = request.Processor
	payment.ProcessorID = request.ProcessorID

	paymentLedger, err := s.accountTransactionRepo.GetTransactionByReferenceIdAndProcessorId(ctxTrx, payment.UUID, request.ProcessorID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentLedger == nil || qrisMetadata.QrType == constant.QrTypeStatic {
		if err := s.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
			Status:  constant.StatusSuccess,
			Channel: constant.ChannelQris,
			Amount:  request.PaidAmount,
		}); err != nil {
			return err
		}

	} else {
		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			TrxDatetime:            payment.TrxDatetime,
			Status:                 constant.StatusSuccess,
			Channel:                constant.ChannelQris,
			Amount:                 request.PaidAmount,
		}
		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}
		if err := s.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			return err
		}
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	// Send callback
	s.sendQrisMpmCallbackOnPaidStatus(ctx, payment, request.PaidAmount)

	return nil
}

func (s *PaymentService) failQrisPayment(ctx context.Context, request *paymentModel.QrisPaymentNotificationRequest, payment *paymentModel.Payment, qrisMetadata *paymentModel.PaymentMetadataQris) error {

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/failQrisPayment")
	defer segment.End()

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

	if qrisMetadata.QrType == constant.QrTypeDynamic {
		if errPayQrMpmDynamic := s.failQrisMpmDynamic(ctxTrx, payment); errPayQrMpmDynamic != nil {
			return errPayQrMpmDynamic
		}
	}
	payment.Processor = request.Processor
	payment.ProcessorID = request.ProcessorID

	paymentLedger, err := s.accountTransactionRepo.GetTransactionByReferenceIdAndProcessorId(ctxTrx, payment.UUID, request.ProcessorID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}
	if paymentLedger == nil || qrisMetadata.QrType == constant.QrTypeStatic {
		if err := s.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
			Status:  constant.StatusFailed,
			Channel: constant.ChannelQris,
			Amount:  request.PaidAmount,
		}); err != nil {
			return err
		}

	} else {
		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			TrxDatetime:            payment.TrxDatetime,
			Status:                 constant.StatusFailed,
			Channel:                constant.ChannelQris,
			Amount:                 request.PaidAmount,
		}
		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}
		if err := s.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			return err
		}
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	// TODO: Product debt - Send callback for failed status

	return nil
}

func (s *PaymentService) payQrisMpmDynamic(
	ctx context.Context, request *paymentModel.QrisPaymentNotificationRequest, payment *paymentModel.Payment) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/payQrisMpmDynamic")
	defer segment.End()

	paidAmountValue, err := decimal.NewFromString(request.PaidAmount.Value)
	if err != nil {
		s.logger.Error(ctx, "error when parsing paid amount", logger.Error(constant.ErrInvalidRequestPayload))
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload)
	}

	// We assume this is full payment, for partial payment will be decided later for the logic
	if paidAmountValue.Cmp(decimal.Zero) > 0 && !paidAmountValue.Equal(payment.Amount) {
		s.logger.Error(ctx, "error payment total amount not match with paid amount", logger.Error(constant.ErrPaymentAmountNotMatch))
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountNotMatch)
	}

	// update status to success
	payment.Status = paymentConstant.PAYMENT_STATUS_SUCCESS
	payment.UpdatedAt = time.Now().UTC()
	if errUpdate := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt); errUpdate != nil {
		return pkgErrors.New(response.HttpErrDatabase, errUpdate)
	}

	return nil
}

// buildQrisMetadata return QRIS Metadata
// you can use the payment.Payment struct and use the GetQRISMetadata() to get same result
func buildQrisMetadata(paymentMetadata *map[string]any) (*paymentModel.PaymentMetadataQris, error) {
	if paymentMetadata == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	jsonData, _ := json.Marshal(paymentMetadata)

	resp := paymentModel.PaymentMetadataQris{}
	_ = json.Unmarshal(jsonData, &resp)

	return &resp, nil
}

func (s *PaymentService) sendQrisMpmCallbackOnPaidStatus(ctx context.Context, payment *paymentModel.Payment, requestAmount commonModel.Amount) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/sendQrisMpmCallbackOnPaidStatus")
	defer segment.End()

	var (
		merchantActor                 = payment.MerchantID
		isSnap                        = true
		paymentCallbackRequestWrapper *anypb.Any
		paymentCallbackEvent          string
		err                           error
	)

	paymentResult := s.buildPaymentForQris(ctx, payment)

	isUnifiedPayment := paymentResult.IsUnifiedPayment
	callbackName := constant.CallbackMasterPaymentSNAPQRIS
	if isUnifiedPayment {
		callbackName = constant.CallbackNamePayment
		isSnap = false
	}

	if isUnifiedPayment {
		paymentCallbackRequestWrapper, err = anypb.New(paymentResult.ToPbUnifiedPaymentCallbackRequest(payment))
		if err != nil {
			s.logger.Error(ctx, "Generate anypb.New ToPbUnifiedPaymentCallbackRequest", logger.Error(err))
			return
		}

		paymentCallbackEvent = fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, payment.Status)
	} else {
		// Get merchant data for merchantName
		// If this request is from submerchant, we use submerchant name instead

		qrisMetadata, errQrisMetadata := buildQrisMetadata(payment.Metadata)
		if errQrisMetadata != nil {
			s.logger.Error(ctx, "Error build QRIS metadata", logger.Error(errQrisMetadata))
			return
		}

		if qrisMetadata.SubMerchantId != "" {
			merchantActor = qrisMetadata.SubMerchantId
		}

		merchant, errMerchant := s.merchantRepo.FindMerchantByID(ctx, merchantActor)
		if errMerchant != nil || merchant == nil {
			s.logger.Error(ctx, "Error find merchant")
			return
		}

		qrRegistration, errFindQr := s.qrisSvc.FindQrRegistrationByExternalID(ctx, merchant.ExternalId)
		if errFindQr != nil {
			s.logger.Error(ctx, "Error find QR registration", logger.Error(errFindQr))
			return
		}

		processorRefNo := ""
		if payment.ProcessorReferenceNumber != nil {
			processorRefNo = *payment.ProcessorReferenceNumber
		}

		referenceID := ""
		if payment.ReferenceID != nil {
			referenceID = *payment.ReferenceID
		}

		builtCallback := &pb.PaymentQrisCallbackRequest{
			OriginalReferenceNo:        payment.UUID,
			OriginalPartnerReferenceNo: referenceID,
			LatestTransactionStatus:    "00", // Success status
			TransactionStatusDesc:      strings.ToLower(paymentConstant.PAYMENT_STATUS_SUCCESS),
			Amount:                     requestAmount.ProtoAmount(),
			AdditionalInfo: &pb.PaymentQrisCallbackAdditionalInfo{
				RRN:             processorRefNo,
				QrType:          qrisMetadata.QrType,
				QrStatus:        constant.QrStatusActive,
				MerchantName:    qrRegistration.MerchantShortName,
				PaymentStatus:   paymentConstant.PAYMENT_STATUS_SUCCESS,
				TransactionDate: util.SnapCompatible(time.Now().UTC()), // TODO: get from snap resp
			},
		}

		if qrisMetadata.QrType == constant.QrTypeDynamic {
			builtCallback.AdditionalInfo.QrExpiredDate = util.SnapCompatible(*payment.ExpiredAt)
			builtCallback.AdditionalInfo.QrStatus = constant.QrStatusInactive
		}

		// Send callback on every payment changes
		paymentCallbackRequestWrapper, err = anypb.New(builtCallback)
		if err != nil {
			s.logger.Error(ctx, "QRIS generate anypb type", logger.Error(err))
			return
		}

		paymentCallbackEvent = constant.CallbackEventPaymentQrisMpmPaid
	}

	merchantID := payment.MerchantID
	// Send callback to merchant that initiated the transaction
	if payment.CreatedBy != nil {
		merchantID = *payment.CreatedBy
	}

	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       callbackName,
		Event:      paymentCallbackEvent,
		MerchantId: merchantID,
		Request:    paymentCallbackRequestWrapper,
		IsSnap:     isSnap,
	}

	_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)

	subject, message := constant.GetNotificationMessage(payment.UUID, payment.Status)

	err = s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, payment.UUID),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   subject,
			Type:      constant.CreateVAPaymentNotifType,
			Message:   message,
			CreatedAt: time.Now().UTC(),
			Status:    payment.Status,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "push notification for payment "+payment.UUID, logger.Error(err))
	}
}

func (s *PaymentService) failQrisMpmDynamic(
	ctx context.Context, payment *paymentModel.Payment) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/failQrisMpmDynamic")
	defer segment.End()

	// update status to void
	payment.Status = paymentConstant.PAYMENT_STATUS_VOID
	payment.UpdatedAt = time.Now().UTC()
	if errUpdate := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt); errUpdate != nil {
		return pkgErrors.New(response.HttpErrDatabase, errUpdate)
	}

	return nil
}
