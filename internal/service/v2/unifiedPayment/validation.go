package unifiedPaymentService

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

var (
	postalCodePattern     = regexp.MustCompile(constant.PostalCodePattern)
	postalCodeHasAlphaNum = regexp.MustCompile(constant.PostalCodeAlphaNumPattern)
)

func (s *UnifiedPaymentService) validateCreatePaymentSession(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/validateCreatePaymentSession")
	defer segment.End()

	// Virtual terminal, card funded payout and auto split (non-authentication) transactions are excluded from this rule, as they permit reuse of the same reference ID
	// and do not apply split routing or order information mechanisms.
	if request.VirtualTerminal != nil || request.CardFundedPayout != nil || (request.AutoSplitPayment != nil && request.AutoSplitPayment.TransactionType != constant.AutoSplitPaymentTypeAuthentication) {
		return nil
	}

	// Validate clientReferenceID
	if paymentByRef, err := s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID, request.ClientReferenceID); err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)

	} else if paymentByRef != nil &&
		!slices.Contains([]string{constant.UnifiedPaymentSessionStatusExpired, constant.UnifiedPaymentSessionStatusCancelled}, paymentByRef.Status) {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrClientReferenceIDAlreadyExist)
	}

	// Validate split routing
	if err = s.validateSplitRouteConfigurations(ctx, request); err != nil {
		return err
	}

	if request.OrderInformation != nil && request.OrderInformation.BillingInformation != nil {
		if err = validatePostalCode(request.OrderInformation.BillingInformation.PostalCode); err != nil {
			return pkgErr.New(response.HttpErrRequest, err)
		}
	}

	return nil
}

func validatePostalCode(postalCode string) error {
	if postalCode == "" {
		return nil
	}

	if len(postalCode) > 10 {
		return constant.ErrPostalCodeTooLong
	}

	// Check first with max length, alphanumeric, and hyphen (-)
	// And check if contains the alphanumeric
	if !postalCodePattern.MatchString(postalCode) || !postalCodeHasAlphaNum.MatchString(postalCode) {
		return constant.ErrPostalCodeInvalidFormat
	}

	return nil
}

func (s *UnifiedPaymentService) validatePaymentActivation(ctx context.Context, merchantID, merchantExternalID, paymentMethod, acquirer string, isSplitRoute bool) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/validatePaymentActivation")
	defer segment.End()

	// Get Payment Method by type and bankCode
	paymentMethod = constant.MapUnifiedPaymentMethod(paymentMethod)
	filterPaymentMethod := paymentModel.GetActivePaymentMethodRequest{
		MerchantID: merchantID,
		Category:   constant.TypePayment,
		Type:       paymentMethod,
	}
	if paymentMethod == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT || paymentMethod == paymentConstant.PAYMENT_METHOD_EWALLET {
		filterPaymentMethod.Acquirer = acquirer
	}

	activePaymentMethod, err := s.paymentMethodSvc.GetActivePaymentMethodDetailForPaymentRequest(ctx, filterPaymentMethod)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)

	} else if activePaymentMethod == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotFound)

	} else if constant.IsDirectPSP(activePaymentMethod.ChannelType) && isSplitRoute {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDoNotApplySplitRouteInFacilitatorModel)
	}

	// Check QRIS acquirer
	if paymentMethod == paymentConstant.PAYMENT_METHOD_QRIS {
		_, errFindQr := s.qrisSvc.FindQrRegistrationByExternalIDAndAcquirer(ctx, merchantExternalID, activePaymentMethod.Acquirer)
		if errFindQr != nil {
			if errFindQr.Error() == pkgErr.New(response.HttpErrNotFound, constant.ErrDataNotFound).Error() {
				return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotRegisteredQR)
			}
			return nil, errFindQr
		}
	}
	return activePaymentMethod, nil
}

func (s *UnifiedPaymentService) validateSplitRouteConfigurations(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) error {
	if !(request.SplitRoutingConfigurations != nil && len(*request.SplitRoutingConfigurations) > 0) {
		return nil
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return pkgErr.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	parentMerchantID := request.MerchantID
	if merchant.ParentID.Valid {
		parentMerchantID = merchant.ParentID.String
	}

	subMerchantIDs, err := s.merchantRepo.GetSubMerchantIdListByParentId(ctx, parentMerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if subMerchantIDs == nil {
		subMerchantIDs = []string{}
	}

	allowedRouteMerchantIDs := subMerchantIDs
	allowedRouteMerchantIDs = append(allowedRouteMerchantIDs, parentMerchantID)

	for _, val := range *request.SplitRoutingConfigurations {
		if !slices.Contains(allowedRouteMerchantIDs, val.MerchantId) || val.MerchantId == request.MerchantID {
			err = errors.New("merchant destination is not allowed")
			s.logger.Warn(ctx, "error validate split routing configurations", logger.Error(err))

			return pkgErr.New(response.HttpErrUnprocessableContent, err)
		}
	}

	return nil
}

func (s *UnifiedPaymentService) validateCard(ctx context.Context, request *unifiedPaymentModel.ValidateCardRequest, paymentMethod *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/validateCard")
	defer segment.End()

	if request.CardPaymentMethodOptions == nil || request.IsRecurringPayment || request.IsVirtualTerminal || request.IsCardFundedPayout || request.IsAutoSplitPayment {
		return nil
	}

	switch request.CardPaymentMethodOptions.ThreeDsMethod {
	case constant.CardThreeDsMethodNever:
		allowByPass3Ds := false
		for _, useCase := range request.CardSupportedUseCases {
			if useCase.AllowBypass3Ds {
				allowByPass3Ds = true
			}
		}

		if !allowByPass3Ds {
			return pkgErr.New(response.HttpErrRequest, constant.ErrPaymentDoesNotAllowedToBypass3Ds)
		}

	case constant.CardThreeDsMethodExternal:
		// Validate merchant allowed to use external 3ds
		allowExternalThreeDs := false
		for _, useCase := range request.CardSupportedUseCases {
			if useCase.AllowExternalThreeDs {
				allowExternalThreeDs = true
			}
		}

		if !allowExternalThreeDs {
			return pkgErr.New(response.HttpErrRequest, constant.ErrExternalThreeDsNotEnabled)
		}

		if request.IsConfirmStep {
			// Validate threeDsInfo object
			if err := s.validateThreeDsInfo(ctx, request.CardPaymentMethodOptions.ThreeDsInfo); err != nil {
				return err
			}
		}
	}

	if request.CardPaymentMethodOptions.CardOnFile != nil {
		if request.CardPaymentMethodOptions.CardOnFile.Initiator == constant.COFInitiatorMerchant {
			if request.CardPaymentMethodOptions.ThreeDsMethod != constant.CardThreeDsMethodNever {
				return pkgErr.New(response.HttpErrRequest, constant.ErrInvalidMITThreeDSMethod)
			}

			if request.CardPaymentMethodOptions.CardOnFile.PreviousNetworkTransactionID == "" {
				return pkgErr.New(response.HttpErrRequest, constant.ErrMissingMerchantPrevioustNetworkTransactionID)
			}
		}

		if request.CardPaymentMethodOptions.CardOnFile.Initiator == constant.COFInitiatorCustomer {
			if request.CardPaymentMethodOptions.CardOnFile.PreviousNetworkTransactionID != "" {
				return pkgErr.New(response.HttpErrRequest, constant.ErrMerchantPreviousNetworkTransactionIDNotAllowedForCIT)
			}
		}
	}

	// Validate processingConfig for per-transaction MID selection
	if err := s.validateProcessingConfig(ctx, request.CardPaymentMethodOptions.ProcessingConfig, paymentMethod, request.CardPaymentMethodOptions.ThreeDsMethod); err != nil {
		return err
	}

	return nil
}

func (s *UnifiedPaymentService) validateProcessingConfig(ctx context.Context,
	processingConfig *unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig,
	paymentMethod *paymentModel.PaymentMethodWithPivot,
	threeDsMethod string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/validateProcessingConfig")
	defer segment.End()

	if processingConfig == nil {
		return nil
	}

	var bankMerchantId string
	var err error

	// resolve by merchant id and get by the merchant tag
	if processingConfig.MerchantIdTag != "" {
		bankMerchantId, err = s.resolveBankMerchantIdFromTag(processingConfig.MerchantIdTag, paymentMethod)
		if err != nil {
			return err
		}
	}

	// if bankMerchantId exist, overwrite this with that value
	if processingConfig.BankMerchantId != "" {
		bankMerchantId = processingConfig.BankMerchantId
	}

	// if it exists, then validate
	if bankMerchantId != "" {
		return s.validateBankMerchantId(ctx, bankMerchantId, threeDsMethod)
	}

	// not exist just pass
	return nil
}

func (s *UnifiedPaymentService) validateBankMerchantId(ctx context.Context, bankMerchantId, threeDsMethod string) error {
	mid, err := s.creditCardProcessorRepo.GetMIDByAcquirerMID(ctx, bankMerchantId)
	if err != nil {
		return pkgErr.New(response.HttpErrValidation, constant.ErrProcessingConfigInvalidMID)
	}

	if mid == nil {
		return pkgErr.New(response.HttpErrValidation, constant.ErrProcessingConfigInvalidMID)
	}

	if threeDsMethod != constant.CardThreeDsMethodExternal && mid.Type != constant.CreditCardMidTypeDirect {
		return pkgErr.New(response.HttpErrValidation, constant.ErrProcessingConfigInvalidMID)
	}

	return nil
}

func (s *UnifiedPaymentService) resolveBankMerchantIdFromTag(merchantIdTag string, paymentMethod *paymentModel.PaymentMethodWithPivot) (string, error) {
	if paymentMethod.MerchantConfigObj == nil ||
		paymentMethod.MerchantConfigObj.PartnerConfig == nil ||
		paymentMethod.MerchantConfigObj.PartnerConfig.Card == nil {
		return "", pkgErr.New(response.HttpErrValidation, constant.ErrProcessingConfigInvalidMID)
	}

	for _, item := range paymentMethod.MerchantConfigObj.PartnerConfig.Card.Items {
		if item.MerchantIDTag == merchantIdTag {
			return item.AcquirerMerchantID, nil
		}
	}

	return "", pkgErr.New(response.HttpErrValidation, constant.ErrProcessingConfigInvalidMID)
}

func (s *UnifiedPaymentService) ValidatePaymentExpiry(ctx context.Context, cmd paymentModel.PaymentRequestExpiryValidation) error {
	validationConfig := paymentModel.PaymentMethodExpiryConfig{}
	expiryRequest := cmd.UnifiedPaymentRequest.ExpiryAt
	isQris := cmd.Method == paymentConstant.PAYMENT_METHOD_QRIS || cmd.Method == constant.UnifiedPaymentMethodQris
	if cmd.Method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.VirtualAccount != nil &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.VirtualAccount.ExpiryAt != nil {
		expiryRequest = util.ValueOfPtr(
			cmd.UnifiedPaymentRequest.PaymentMethodOptions.VirtualAccount.ExpiryAt,
		)
	}

	if (cmd.Method == paymentConstant.PAYMENT_METHOD_QRIS || cmd.Method == constant.UnifiedPaymentMethodQris) &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.QR != nil &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.QR.ExpiryAt != nil {
		expiryRequest = util.ValueOfPtr(
			cmd.UnifiedPaymentRequest.PaymentMethodOptions.QR.ExpiryAt,
		)
	}

	// card have different payment method name for Unified Payment
	isCard := cmd.Method == paymentConstant.PAYMENT_METHOD_CREDIT_CARD || cmd.Method == constant.UnifiedPaymentMethodCard
	if isCard &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.Card != nil &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.Card.ExpiryAt != nil {
		expiryRequest = util.ValueOfPtr(
			cmd.UnifiedPaymentRequest.PaymentMethodOptions.Card.ExpiryAt,
		)
	}

	if cmd.Method == paymentConstant.PAYMENT_METHOD_EWALLET &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.Ewallet != nil &&
		cmd.UnifiedPaymentRequest.PaymentMethodOptions.Ewallet.ExpiryAt != nil {
		expiryRequest = util.ValueOfPtr(
			cmd.UnifiedPaymentRequest.PaymentMethodOptions.Ewallet.ExpiryAt,
		)
	}

	if expiryRequest.IsZero() {
		return nil
	}

	paymentMethodList := []string{
		paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		paymentConstant.PAYMENT_METHOD_QRIS,
		constant.UnifiedPaymentMethodQris,
		paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
		constant.UnifiedPaymentMethodCard,
		paymentConstant.PAYMENT_METHOD_EWALLET,
	}

	if !slices.Contains(paymentMethodList, cmd.Method) {
		return nil
	}

	// set default config for virtual account
	if cmd.Method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT && s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil {
		vaConfig := s.config.UnifiedPaymentConfig.VirtualAccountConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: vaConfig.MaxExpiryDuration,
			Unit:     vaConfig.MaxExpiryDurationUnit,
		}
	}

	// set default config for qris
	if isQris && s.config.UnifiedPaymentConfig.QrConfig != nil {
		qrConfig := s.config.UnifiedPaymentConfig.QrConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: qrConfig.MaxExpiryDuration,
			Unit:     qrConfig.MaxExpiryDurationUnit,
		}
	}

	// set default config for cards
	if isCard && s.config.UnifiedPaymentConfig.CardConfig != nil {
		cardConfig := s.config.UnifiedPaymentConfig.CardConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: cardConfig.MaxExpiryDuration,
			Unit:     cardConfig.MaxExpiryDurationUnit,
		}
	}

	// set default config for ewallet
	if cmd.Method == paymentConstant.PAYMENT_METHOD_EWALLET && s.config.UnifiedPaymentConfig.EwalletConfig != nil {
		ewalletConfig := s.config.UnifiedPaymentConfig.EwalletConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: ewalletConfig.MaxExpiryDuration,
			Unit:     ewalletConfig.MaxExpiryDurationUnit,
		}
	}

	if s.config.UnifiedPaymentConfig.ExpiryConfig != nil && !s.config.UnifiedPaymentConfig.ExpiryConfig.ShouldValidateExpiry(cmd.MerchantID) {
		return nil
	}

	// replace config from database
	if cmd.PaymentMethod.PaymentMethod.ConfigObj != nil &&
		cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Unit != "" &&
		cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Duration > 0 {

		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Duration,
			Unit:     cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Unit,
		}
	}

	maxValidationTime := validationConfig.ToDateTime()
	// return error if expiry request is greater than max validation time
	if !maxValidationTime.IsZero() && expiryRequest.After(maxValidationTime) {
		if cmd.Request != nil && cmd.Request.IsSnap {
			expiryField := "expiredDate"
			if cmd.Request.PaymentMethod == paymentConstant.PAYMENT_METHOD_QRIS {
				expiryField = "validityPeriod"
			}
			return pkgErr.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, expiryField))
		}

		maxValidationTimeStr := fmt.Sprintf("%d %s", validationConfig.Duration, validationConfig.Unit)
		return pkgErr.New(response.HttpErrRequest, fmt.Errorf(constant.ErrExceedMaxExpiryDate, cmd.Method, maxValidationTimeStr))
	}

	return nil
}

// validateThreeDsInfo validates the 3DS information for external 3DS flow
func (s *UnifiedPaymentService) validateThreeDsInfo(ctx context.Context, threeDsInfo *unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/validateThreeDsInfo")
	defer segment.End()

	// Validate that threeDsInfo is required for threeDsMethod = EXTERNAL
	if threeDsInfo == nil {
		return pkgErr.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRequireThreeDsInfoForThreeDsMethodExternal)
	}

	// Reject 3DS v1 payloads (version starts with "1.")
	if len(threeDsInfo.ThreeDSVersion) >= 2 && threeDsInfo.ThreeDSVersion[0] == '1' && threeDsInfo.ThreeDSVersion[1] == '.' {
		return pkgErr.New(response.HttpErrRequest, constant.ErrUnifiedPaymentInvalidThreeDsInfoFormat)
	}

	// Validate transactionStatus = Y (authentication success)
	if threeDsInfo.TransactionStatus != "Y" {
		return pkgErr.New(response.HttpErrRequest, constant.ErrUnifiedPayment3DsAuthenticationNotSuccessful)
	}

	// Validate ECI matches authenticationScheme
	if err := s.validateECIMatchesScheme(threeDsInfo.ECI, threeDsInfo.AuthenticationScheme); err != nil {
		return err
	}

	return nil
}

// validateECIMatchesScheme validates that the ECI value matches the authentication scheme
func (s *UnifiedPaymentService) validateECIMatchesScheme(eci string, scheme string) error {
	schemeECIMap := map[string]map[string]bool{
		"VISA":       {"05": true},
		"MASTERCARD": {"02": true},
		"JCB":        {"05": true},
		"AMEX":       {"05": true},
		"UNIONPAY":   {"05": true, "02": true},
	}

	allowedECIs, schemeExists := schemeECIMap[scheme]
	if !schemeExists {
		return pkgErr.New(response.HttpErrRequest, constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult)
	}

	if !allowedECIs[eci] {
		return pkgErr.New(response.HttpErrRequest, constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult)
	}

	return nil
}
