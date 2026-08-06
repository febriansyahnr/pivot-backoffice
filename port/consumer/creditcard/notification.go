package creditcard

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CreditCardConsumer) PaymentNotification(ctx context.Context, body []byte, channel string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/creditcard/PaymentNotification")
	defer segment.End()

	var (
		req              creditcardModel.CardPaymentNotificationRequest
		metricAttributes = map[string]any{}
	)

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from PaymentNotification", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	err = json.Unmarshal(body, &req)
	if err != nil {
		return pkgError.New(response.HttpErrRequest, err)
	}

	c.logger.Info(ctx, "recieve CC payment notification", logger.String("body", string(body)))

	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentPaymentProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgError.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "error when record processed payment custom metric", logger.Error(errMetric))
		}
	}()

	metricAttributes["status"] = req.Data.PaymentStatus
	metricAttributes["paymentMethod"] = constant.UnifiedPaymentMethodCard

	// Update payment metadata when ChargeStatus is WAITING_FOR_EXTERNAL_FDS
	if req.Data.PaymentStatus == constant.ChargeStatusWaitingForExternalFDS {
		if err = c.creditcardSvc.PaymentNotificationFDS(ctx, req); err != nil {
			return err
		}

		return nil
	}

	if processedUnifiedPaymentV2, err := c.checkAndProcessUnifiedPaymentV2(ctx, req.Data.PaymentUUID.String(),
		req.Data,
		&unifiedPaymentModel.PaymentNotificationRequest{
			PaymentMethodType: constant.UnifiedPaymentMethodCard,
			ChargeStatus:      constant.MapProcessorToChargeStatus(req.Data.PaymentStatus, req.Data.Type),
			Amount: unifiedPaymentModel.Amount{
				Currency: req.Data.Currency,
				Value:    req.Data.Amount.InexactFloat64(),
			},
			TrxDatetime:            req.Data.Updated,
			Processor:              constant.CreditCardCoreProcessor,
			ProcessorID:            req.Data.AcquirerTransactionID,
			ProcessorTransactionID: req.Data.TransactionID.String(),
		}); err != nil {
		return err
	} else if processedUnifiedPaymentV2 {
		return nil
	}

	// Call CallbackPayment
	if err = c.creditcardSvc.PaymentNotification(ctx, req); err != nil {
		return err
	}

	return nil
}

func (c *CreditCardConsumer) checkAndProcessUnifiedPaymentV2(
	ctx context.Context,
	paymentSessionID string,
	paymentNotificationData creditcardModel.PaymentNotificationDataRequest,
	request *unifiedPaymentModel.PaymentNotificationRequest,
) (isProcessed bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/creditcard/PaymentNotification")
	defer segment.End()

	// Check the payment is unified payment v2 or not
	isUnifiedPaymentV2 := false
	payment, errFind := c.paymentSvc.GetDetailByID(ctx, paymentSessionID)
	if errFind != nil {
		return isProcessed, errFind
	}

	paymentMetadata := map[string]interface{}{}
	if payment.Metadata != nil {
		paymentMetadata = *payment.Metadata
	}

	if isUnifiedPayment, exists := paymentMetadata["isUnifiedPaymentV2"]; exists {
		isUnifiedPaymentV2 = isUnifiedPayment.(bool)
	}

	if slices.Contains([]string{constant.CreditCardStatusVoid, constant.CreditCardStatusRefunded}, paymentNotificationData.PaymentStatus) {
		totalRefundedAmount, err := c.refundSvc.GetTotalRefundedAmount(ctx, paymentSessionID)
		if err != nil {
			c.logger.Error(ctx, "error retrieve payment refunds", logger.String("paymentSessionId", paymentSessionID))
			return false, err
		}
		if totalRefundedAmount > 0 {
			// Skip payment notification
			c.logger.Info(ctx, "skip CC payment notification for unified refund case")
			return true, nil
		}

		var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
		metadataB, _ := json.Marshal(payment.Metadata)
		_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

		if util.ValueOfPtr(unifiedPaymentMetadata.PaymentMethodOptions.Card).CaptureMethod == constant.UnifiedPaymentCardCaptureMethodManual {
			// Skip payment notification
			c.logger.Info(ctx, "skip CC payment notification for capture method MANUAL")
			return true, nil
		}
	}

	if isUnifiedPaymentV2 {
		// Skip capture notification, because it is already handled in processCapture
		if paymentNotificationData.Type == constant.CardTransactionTypeCapture {
			// Skip payment notification
			c.logger.Info(ctx, "skip CC payment notification for transaction type is CAPTURE", logger.String("paymentId", paymentNotificationData.PaymentUUID.String()))
			return true, nil
		}

		// Force Success For Authorized One Dollar Authorization
		if payment.Type == constant.UnifiedPaymentOneDollarAuthorization &&
			paymentNotificationData.Type == constant.CardTransactionTypeAuthorization &&
			request.ChargeStatus == constant.ChargeStatusWaitingForCapture {

			request.ChargeStatus = constant.ChargeStatusSuccess
		}

		charge, err := c.orchestratorSvc.FindByReference(ctx, paymentSessionID, constant.TypePayment)
		if err != nil {
			return isProcessed, err
		}

		if charge == nil {
			c.logger.Info(ctx, "payment charge not found", logger.String("paymentID", payment.UUID))
			return isProcessed, pkgError.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
		}

		chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
		_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
			MethodDetail interface{} `json:"methodDetail"`
		}{
			MethodDetail: chargeMethodDetails,
		})

		// Pass paymentSessionId and chargeId
		request.PaymentSessionID = paymentSessionID
		request.ChargeID = charge.UUID.String()
		request.ChargePaymentMethodDetails = &unifiedPaymentModel.ChargePaymentMethodDetails{
			Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{},
		}

		if chargeMethodDetails.Card != nil {
			request.ChargePaymentMethodDetails.Card = chargeMethodDetails.Card
			request.ChargePaymentMethodDetails.Card.Description = paymentNotificationData.Description
			request.ChargePaymentMethodDetails.Card.SettlementDate = paymentNotificationData.SettlementDate
			request.ChargePaymentMethodDetails.Card.MerchantCategoryCode = paymentNotificationData.MerchantCategoryCode
		}
		if paymentNotificationData.Device != nil {
			request.ChargePaymentMethodDetails.Card.Device = &unifiedPaymentModel.ChargePaymentMethodDetailCardDevice{
				Browser:          paymentNotificationData.Device.Browser,
				IPAddress:        paymentNotificationData.Device.IPAddress,
				MobilePhoneModel: paymentNotificationData.Device.MobilePhoneModel,
			}
		}
		if paymentNotificationData.Error != nil {
			request.ChargePaymentMethodDetails.Card.Error = &unifiedPaymentModel.ChargePaymentMethodDetailCardError{
				Cause:       paymentNotificationData.Error.Cause,
				Explanation: paymentNotificationData.Error.Explanation,
			}
		}

		if paymentNotificationData.MIDInfo != nil {
			request.ChargePaymentMethodDetails.Card.MIDInfo = &unifiedPaymentModel.MIDInfo{
				MID:  paymentNotificationData.MIDInfo.MID,
				Type: paymentNotificationData.MIDInfo.Type,
			}
		}

		if paymentNotificationData.CardData != nil {
			first6 := paymentNotificationData.CardData.First8Digit
			if len(first6) > 6 {
				first6 = first6[:6]
			}

			request.ChargePaymentMethodDetails.Card.First6 = first6
			request.ChargePaymentMethodDetails.Card.First8 = paymentNotificationData.CardData.First8Digit
			request.ChargePaymentMethodDetails.Card.Last4 = paymentNotificationData.CardData.Last4Digit
			request.ChargePaymentMethodDetails.Card.ExpMonth = types.String(paymentNotificationData.CardData.ExpiryMonth)
			request.ChargePaymentMethodDetails.Card.ExpYear = types.String(paymentNotificationData.CardData.ExpiryYear)
			request.ChargePaymentMethodDetails.Card.Fingerprint = paymentNotificationData.CardData.Fingerprint
			request.ChargePaymentMethodDetails.Card.BankMerchantID = paymentNotificationData.BankMerchantID
			request.ChargePaymentMethodDetails.Card.SaveForFutureUse = util.ValueToPtr(paymentNotificationData.CardData.SavedFutureUse)
			request.ChargePaymentMethodDetails.Card.CardHolderName = paymentNotificationData.CardData.CardHolderName
			request.ChargePaymentMethodDetails.Card.CardName = paymentNotificationData.CardData.CardName

			request.ChargePaymentMethodDetails.Card.BinInformations = unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
				Type:        paymentNotificationData.CardData.CardType,
				IssuingBank: paymentNotificationData.CardData.CardIssuing,
				Brand:       paymentNotificationData.CardData.CardBrand,
				Country:     paymentNotificationData.CardData.IssuingCountry,
			}
		}

		if paymentNotificationData.AuthenticationData != nil {
			request.ChargePaymentMethodDetails.Card.AuthenticationResult = &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
				ThreeDsVersion:        paymentNotificationData.AuthenticationData.ThreeDsVer,
				ThreeDsResult:         paymentNotificationData.AuthenticationData.AuthenticationResult,
				ThreeDsMethod:         paymentNotificationData.AuthenticationMethod,
				EciCode:               paymentNotificationData.AuthenticationData.EciCode,
				TransactionID:         paymentNotificationData.AuthenticationData.TransactionID,
				TransactionStatus:     paymentNotificationData.AuthenticationData.TransactionStatus,
				AuthenticationScheme:  paymentNotificationData.AuthenticationData.AuthenticationScheme,
				AcsTransactionID:      paymentNotificationData.AuthenticationData.AcsTransactionID,
				AcsReference:          paymentNotificationData.AuthenticationData.AcsReference,
				AuthenticationTime:    paymentNotificationData.AuthenticationData.AuthenticationTime,
				CallbackTransactionID: paymentNotificationData.AuthenticationData.CallbackTransactionID,
			}
		}

		if paymentNotificationData.AuthorizationData != nil {
			retrievalReferenceNumber := paymentNotificationData.AuthorizationData.TransactionReference
			if paymentNotificationData.AuthorizationData.RetrievalReferenceNumber != "" {
				retrievalReferenceNumber = paymentNotificationData.AuthorizationData.RetrievalReferenceNumber
			}

			request.ChargePaymentMethodDetails.Card.AuthorizationResult = &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
				AcquirerReferenceNumber:  paymentNotificationData.AuthorizationData.AcquirerTransactionID,
				RetrievalReferenceNumber: retrievalReferenceNumber,
				Stan:                     paymentNotificationData.AuthorizationData.Stan,
				AvsResult:                paymentNotificationData.AuthorizationData.AvsResult,
				CvvResult:                paymentNotificationData.AuthorizationData.CvvResult,
				AuthorizedAmount: unifiedPaymentModel.Amount{
					Currency: paymentNotificationData.Currency,
					Value:    paymentNotificationData.Amount.InexactFloat64(),
				},
				IssuerAuthorizationCode: paymentNotificationData.AuthorizationData.AcquirerResponseCode,
				AuthorizationID:         paymentNotificationData.AuthorizationData.AuthorizationID,
				MerchantAdviceCode:      paymentNotificationData.AuthorizationData.MerchantAdviceCode,
				NetworkTransactionID:    paymentNotificationData.AuthorizationData.AcquirerTransactionID,
			}
			request.ChargePaymentMethodDetails.Card.ApprovalCode = paymentNotificationData.AuthorizationData.ApprovalCode
		}
		if paymentNotificationData.ResponseCode != nil {
			request.ChargePaymentMethodDetails.Card.ResponseCode = &unifiedPaymentModel.ChargePaymentMethodDetailCardResponseCode{
				AcquirerCode:          paymentNotificationData.ResponseCode.AcquirerCode,
				AcquirerMessage:       paymentNotificationData.ResponseCode.AcquirerMessage,
				GatewayCode:           paymentNotificationData.ResponseCode.GatewayCode,
				GatewayRecommendation: paymentNotificationData.ResponseCode.GatewayRecommendation,
			}
		}

		// Process unified payment v2
		if err = c.unifiedPaymentSvc.ProcessNotification(ctx, request); err != nil {
			return isProcessed, err
		}

		isProcessed = true
	}

	return
}
