package unifiedPaymentService

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// GetFailureCodeOfMethodDetail derives a unified failure code for card payments based on
// transaction status, gateway response, authentication result, and issuer authorization code.
// It returns an empty string when trxStatus is not constant.StatusFailed, methodDetail is nil,
// or no card detail is present.
func (s *UnifiedPaymentService) GetFailureCodeOfMethodDetail(trxStatus string, methodDetail *unifiedPaymentModel.ChargePaymentMethodDetails) string {
	if trxStatus != constant.StatusFailed || methodDetail == nil || methodDetail.Card == nil {
		return ""
	}

	// cannot using return early approach due to incosistent
	// method detail data. in this case mpgs and card simulation have different behavior
	failureCode := constant.FailureCodeUnknown

	// Check gateway response code for cancellation
	if methodDetail.Card.ResponseCode != nil && methodDetail.Card.ResponseCode.GatewayCode == constant.CreditCardGatewayCodeAborted {
		failureCode = constant.FailureCodeCancelledByUser
	} else if methodDetail.Card.AuthenticationResult != nil && methodDetail.Card.AuthenticationResult.ThreeDsResult != constant.CreditCardAuthenticationSuccess {
		failureCode = constant.FailureCodeAuthenticationFailed
	}

	// Check issuer authorization code
	issuerAuthorizationCode := ""
	if methodDetail.Card.AuthorizationResult != nil {
		issuerAuthorizationCode = methodDetail.Card.AuthorizationResult.IssuerAuthorizationCode
	}

	switch issuerAuthorizationCode {
	case "01", "03", "05", "06", "12", "13", "22", "40", "57", "61", "62", "63", "64", "65", "6P", "70", "82", "92", "93", "100", "109", "110", "115", "54", "101", "N7":
		failureCode = constant.FailureCodeDeclinedByChannel
	case "14", "15", "21", "46", "52", "53", "78", "79", "111", "04", "07", "41", "43", "200", "34", "59", "83":
		failureCode = constant.FailureCodeSuspectedFraud
	case "51", "116", "121":
		failureCode = constant.FailureCodeInsufficientFund
	case "19", "80", "90", "91", "96", "911":
		failureCode = constant.FailureCodeChannelUnavailable
	}

	return failureCode
}

func (s *UnifiedPaymentService) isMerchantExcludedToSendSurnameOnCallback(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/isMerchantExcludedToSendSurnameOnCallback")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagMerchantExcludedSendSurname, attr, false)
	return enabled
}

// overridePartnerErrorToRateLimitIfNeeded transforms internal server errors (5xx) into rate limit errors (429)
// when the request originates from the checkout UI feature. This provides a better user experience by indicating
// the request should be retried rather than displaying a generic server error.
func (s *UnifiedPaymentService) overridePartnerErrorToRateLimitIfNeeded(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	if ctx.Value(constant.CtxFeatureName) == nil {
		return err
	}

	httpErrorType, originErr := pkgErr.ExtractError(err)
	confirmedFromCheckoutUI := ctx.Value(constant.CtxFeatureName).(string) == constant.FeatureConfirmPaymentCheckoutUI

	if !confirmedFromCheckoutUI {
		return err
	}

	// TBD: product have plan to apply it into OPEN API
	// Override 5xx (internal server errors) to 429 rate limit
	switch httpErrorType {
	case httpResponse.HttpErrInternal, httpResponse.HttpErrRequestTimeout:
		return pkgErr.New(httpResponse.HttpErrRequestLimitExceeded, originErr)
	}

	return err
}

func (s *UnifiedPaymentService) GetCardMIDAcquirer(ctx context.Context, payment *paymentModel.Payment, notifReq *unifiedPaymentModel.PaymentNotificationRequest) (string, error) {
	paymentMethod, err := s.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, payment.PaymentMethodID, payment.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant payment method", logger.Error(err), logger.String("paymentID", payment.UUID))
		return "", err
	}

	if paymentMethod == nil || paymentMethod.MerchantConfigObj == nil {
		s.logger.Warn(ctx, "payment method not found", logger.String("paymentID", payment.UUID))
		return "", nil
	}

	return paymentMethod.MerchantConfigObj.GetCardAcquirer(notifReq.ChargePaymentMethodDetails.GetCardMID()), nil
}

func (s *UnifiedPaymentService) PrepareCardAuthentication(ctx context.Context, cardAuthenticationRequest *unifiedPaymentModel.CardAuthenticationRequest) (context.Context, creditcardCoreProcessorModel.AuthenticationRequest, error) {
	var authRequest creditcardCoreProcessorModel.AuthenticationRequest
	fnCtx := context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "Unified-Payment",
		ReferenceId: cardAuthenticationRequest.MerchantID,
		OriginId:    cardAuthenticationRequest.PaymentID,
	})

	certPEM, err := s.creditcardSvc.GetCardEncryptionPublicKey(fnCtx, cardAuthenticationRequest.MerchantID)
	if err != nil {
		return fnCtx, authRequest, err
	}

	payloadBytes, _ := json.Marshal(cardAuthenticationRequest)

	s.logger.Info(fnCtx, "Processing payment authentication "+cardAuthenticationRequest.PaymentID, logger.Any("detail", cardAuthenticationRequest))
	encryptedPayload, err := s.cryptoProvider.EncryptPKCS7(certPEM, payloadBytes)
	if err != nil {
		s.logger.Error(fnCtx, "Failed when encrypt request payload for card authentication", logger.Error(err))
		return fnCtx, authRequest, err
	}

	authRequest = creditcardCoreProcessorModel.AuthenticationRequest{
		MerchantID:       cardAuthenticationRequest.MerchantID,
		PaymentID:        cardAuthenticationRequest.PaymentID,
		EncryptedPayload: encryptedPayload,
	}

	return fnCtx, authRequest, nil
}
