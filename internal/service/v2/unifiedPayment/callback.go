package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	protoCommon "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

func (s *UnifiedPaymentService) SendCallback(ctx context.Context, payment *paymentModel.Payment) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/sendCallback")
	defer segment.End()

	if s.shouldSkipSendCallback(payment) {
		if payment.Type == constant.UnifiedPaymentOneDollarAuthorization {
			unifiedPaymentMetadata := payment.ToUnifiedPaymentMetadata()

			if unifiedPaymentMetadata != nil &&
				unifiedPaymentMetadata.OneDollarAuthorization != nil &&
				unifiedPaymentMetadata.OneDollarAuthorization.UseCase == constant.UnifiedPaymentUseCaseCardFundedPayoutSavedCards {

				// Publish stomp notification
				s.sendStompNotification(ctx, payment, util.ValueOfPtr(payment.ReferenceID))
			}
		}

		s.logger.Info(ctx, fmt.Sprintf("Payment callback notifications are disabled for %s transactions", util.ToTitle(payment.Type)), logger.Any("payment", payment))
		return
	}

	var (
		paymentCallbackRequestWrapper *anypb.Any
		paymentCallbackEvent          string
		customerInfo                  *unifiedPaymentModel.CustomerInformationResponse
	)

	charge, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "[UnifiedPaymentV2] Failed to send callback due to error get charge by reference", logger.Error(err))
		return
	} else if charge == nil {
		return
	}

	if payment.CustomerID != "" {
		customer, err := s.customerRepo.GetCustomerById(ctx, payment.CustomerID, payment.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "[UnifiedPaymentV2] Failed to send callback due to error get customer", logger.Error(err))
			return
		}

		if customer != nil {
			customerInfo = customer.ToUnifiedPaymentCustomerResponse()

			// force nil for omitempty
			if s.isMerchantExcludedToSendSurnameOnCallback(ctx, customer.MerchantID) {
				customerInfo.Surname = nil
			}
		}
	}

	merchantID := payment.MerchantID
	// Send callback to merchant that initiated the transaction
	if payment.CreatedBy != nil && payment.CreatedFrom != constant.SourceMerchantPortal {
		// Merchant portal created by is populated with user id
		merchantID = *payment.CreatedBy
	}

	callbackName := constant.CallbackNamePayment

	isSnap, _ := util.LookupMapAnyByKey[bool](payment.Metadata, "isSnap")
	if isSnap {
		if payment.PaymentMethod.Type == paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
			s.logger.Info(ctx, "The delivery of the expired payment callback for SNAP VA was ignored", logger.String("paymentId", payment.UUID), logger.Any("clientReferenceId", payment.ReferenceID))
			return
		}

		callbackName = constant.CallbackMasterPaymentSNAPQRIS

		referenceID := ""
		if payment.ReferenceID != nil {
			referenceID = *payment.ReferenceID
		}
		snapQrisCallback := &pb.PaymentQrisCallbackRequest{
			OriginalReferenceNo:        payment.UUID,
			OriginalPartnerReferenceNo: referenceID,
			LatestTransactionStatus:    "05", // Canceled
			TransactionStatusDesc:      "expired",
			Amount: &protoCommon.Amount{
				Currency: payment.Currency,
				Value:    fmt.Sprintf("%.02f", payment.Amount.InexactFloat64()),
			},
		}
		paymentCallbackRequestWrapper, err = anypb.New(snapQrisCallback)

	} else {
		paymentCallbackRequestWrapper, err = anypb.New(payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customerInfo))
	}
	if err != nil {
		s.logger.Error(ctx, "Generate anypb.New ToPbUnifiedPaymentCallbackRequest", logger.Error(err))
		return
	}

	if s.isPaymentMigrationV1ToV2Enabled(ctx, merchantID) {
		var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
		metadataB, _ := json.Marshal(payment.Metadata)
		_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
		if unifiedPaymentMetadata.IsMigratingFromV1 {
			paymentResponse, err := s.paymentSvc.FindPaymentById(ctx, payment.UUID, payment.MerchantID)
			if err != nil {
				s.logger.Error(ctx, "[UnifiedPaymentV2] Failed to find payment by id", logger.Error(err))
				return
			}

			payment.Status = constant.MapUnifiedPaymentStatusToV1(payment.Status)
			switch constant.MapToUnifiedPaymentMethod(payment.PaymentMethod.Type) {
			case constant.UnifiedPaymentMethodQris:
				paymentResponse.BuildQRISDataFromPaymentV2(payment, charge)
			case constant.UnifiedPaymentMethodVA:
				paymentResponse.BuildQVADataFromPaymentV2(payment, charge)
			case constant.UnifiedPaymentMethodCard:
				payment.BuildCardDataFromPaymentV2(charge)
			}

			paymentCallbackRequestWrapper, err = anypb.New(paymentResponse.ToPbUnifiedPaymentCallbackRequest(payment))
			if err != nil {
				s.logger.Error(ctx, "Generate anypb.New ToPbUnifiedPaymentCallbackRequest", logger.Error(err))
				return
			}
		}
	}

	paymentCallbackEvent = fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, payment.Status)

	// publish callback
	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       callbackName,
		Event:      paymentCallbackEvent,
		MerchantId: merchantID,
		Request:    paymentCallbackRequestWrapper,
		IsSnap:     isSnap,
	}
	_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)

	s.sendStompNotification(ctx, payment, payment.UUID)
}

func (s *UnifiedPaymentService) sendPaymentChargeCallback(ctx context.Context, chargeID string, payment *paymentModel.Payment) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/sendPaymentChargeCallback")
	defer segment.End()

	accountTrx, err := s.accountTransactionRepo.FindByID(ctx, chargeID)
	if err != nil || accountTrx == nil {
		s.logger.Error(ctx, "[UnifiedPaymentV2] Failed to find charge by id", logger.Error(err))
		return
	}
	chargeResponse := unifiedPaymentModel.AccountTransactionToChargeResponse(accountTrx)
	if slices.Contains([]string{constant.ChargeStatusSuccess, constant.ChargeStatusWaitingForCapture}, chargeResponse.Status) {
		chargeResponse.SetAuthorizedAmount(&unifiedPaymentModel.Amount{
			Currency: accountTrx.Currency,
			Value:    payment.Amount.InexactFloat64(),
		})
	}

	merchantID := accountTrx.MerchantID.String()
	// Send callback to merchant that initiated the transaction
	if payment.CreatedBy != nil {
		merchantID = *payment.CreatedBy
	}

	// Set captureHistories to the chargeResp if any & not excluded ff
	if !s.isMerchantExcludedToSendCaptureHistoryOnCallback(ctx, merchantID) {
		chargeResponse.SetCaptureHistories(payment.PaymentCaptures)
	}

	// unifiedMetadata from payment
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	if unifiedPaymentMetadata.MethodDetail != nil {
		if unifiedPaymentMetadata.MethodDetail.Qr != nil {
			chargeResponse.Qr = unifiedPaymentMetadata.MethodDetail.Qr
		}

		if unifiedPaymentMetadata.MethodDetail.VirtualAccount != nil {
			chargeResponse.VirtualAccount = unifiedPaymentMetadata.MethodDetail.VirtualAccount

			if unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount != nil {
				chargeResponse.VirtualAccount.VirtualAccountTrxType = unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.VirtualAccountTrxType
				if payment.BankReferenceId != "" && unifiedPaymentMetadata.IsSnap != nil && *unifiedPaymentMetadata.IsSnap {
					chargeResponse.VirtualAccount.BankReferenceNo = payment.BankReferenceId
				}
			}
		}
	}

	callbackName := constant.CallbackNamePayment

	chargeCallbackRequestWrapper, err := anypb.New(chargeResponse.ToPbChargeResponse())
	if err != nil {
		s.logger.Error(ctx, "Generate anypb.New ToPbUnifiedPaymentCallbackRequest", logger.Error(err))
		return
	}

	paymentCallbackEvent := fmt.Sprintf(constant.CallbackEventUnifiedPaymentChargePattern, chargeResponse.Status)

	isSnap := false
	if unifiedPaymentMetadata.IsSnap != nil && *unifiedPaymentMetadata.IsSnap {
		isSnap = true
		if chargeResponse.Qr != nil {
			callbackName = constant.CallbackMasterPaymentSNAPQRIS
			paymentCallbackEvent = constant.CallbackEventPaymentQrisMpmPaid
		} else if chargeResponse.VirtualAccount != nil {
			callbackName = constant.CallbackMasterPaymentSNAPVA
			paymentCallbackEvent = constant.CallbackEventPaymentVirtualAccountPaid
		}
	}

	// publish callback
	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       callbackName,
		Event:      paymentCallbackEvent,
		MerchantId: merchantID,
		Request:    chargeCallbackRequestWrapper,
		IsSnap:     isSnap,
	}
	_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)

	// Publish to payment UI
	s.sendStompNotification(ctx, payment, payment.UUID)
}

func (s *UnifiedPaymentService) sendStompNotification(ctx context.Context, payment *paymentModel.Payment, identifier string) {
	subject, message := constant.GetNotificationMessage(identifier, payment.Status)
	err := s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, identifier),
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

func (s *UnifiedPaymentService) shouldSkipSendCallback(payment *paymentModel.Payment) bool {
	return slices.Contains([]string{constant.TypeVirtualTerminal, constant.TypeCardFundedPayout, constant.UnifiedPaymentOneDollarAuthorization}, payment.Type) ||
		payment.IsAutoSplitSubPayments() ||
		(payment.IsAutoSplitPaymentAuth() && !s.isAutoSplitPaymentAllowedSendCallback(payment))
}

// isAutoSplitPaymentAllowedSendCallback determines whether a callback should be sent
// for a payment that uses auto-split payment. It returns true only when:
//   - the payment has an AutoSplitPayment attached,
//   - the payment is authentication type and
//   - either the payment is still in processing status or the auto-split summary is not PROCESSING
func (s *UnifiedPaymentService) isAutoSplitPaymentAllowedSendCallback(payment *paymentModel.Payment) bool {
	if payment.AutoSplitPayment == nil || payment.IsAutoSplitSubPayments() {
		return false
	}

	return payment.IsAutoSplitPaymentAuth() &&
		(payment.Status == constant.UnifiedPaymentSessionStatusProcessing ||
			payment.AutoSplitPayment.Summary.Status != constant.AutoSplitPaymentStatusProcessing)
}
