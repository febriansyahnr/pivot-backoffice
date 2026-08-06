package v2InternalUnifiedPaymentController

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (c *paymentController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/Confirm")
	defer segment.End()

	var (
		err  error
		resp *unifiedPaymentModel.UnifiedPaymentSessionResponse
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	payload := unifiedPaymentModel.NewCreateUnifiedPaymentSessionRequest()
	payload.MerchantID = merchantID
	payload.ParentMerchantID = merchantAuth.MerchantId
	payload.CreatedBy = merchantAuth.MerchantId

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	ffContext := ffcontext.NewEvaluationContext(c.config.Environment)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, payload.MerchantID)
	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyUnifiedPaymentCustomerAndOrderObjectEligibleMerchant, ffContext, true)
	if !enabled {
		payload.OrderInformation = nil
		payload.CustomerID = ""
		payload.CustomerInformation = nil
	}

	if err = c.validate.Struct(payload); err != nil {
		c.logger.Warn(ctx, "create session validation error", logger.Error(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Set expiry at if not set
	payload.SetDefaultExpiryAtIfNotSet()

	// Simulate api error
	if err = c.simulateApiError(ctx, payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	// Validate
	if errValidate := c.validatePayload(payload); errValidate != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errValidate)
		return
	}

	err = c.ValidateCustomerPayload(ctx, payload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	// Call service to create unified payment
	if resp, err = c.unifiedPaymentSvc.CreateSession(ctx, payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *paymentController) validatePayload(payload *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) error {
	if (payload.PaymentType == constant.UnifiedPaymentTypeSingle || !payload.ExpiryAt.IsZero()) && payload.ExpiryAt.Before(time.Now().UTC()) {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("expiry time is not permitted to be less than current time"))
	}

	if payload.Amount.Currency == constant.CurrencyIDR && math.Mod(payload.Amount.Value, 1) != 0 {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("amount value is not permitted to use decimal format"))
	}

	if payload.RecurringID == "" && payload.AutoConfirm && payload.PaymentMethod == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrConfirmShouldChoosePaymentMethod)
	}

	if payload.PaymentType == constant.UnifiedPaymentTypeMultiple {
		if payload.Mode != constant.UnifiedPaymentModeAPI {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedForNonAPI)
		}

		if !payload.AutoConfirm {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedAutoConfirmFalse)
		}
	}

	paymentMethodType := ""
	if payload.PaymentMethod != nil {
		paymentMethodType = payload.PaymentMethod.Type
	}

	if payload.SaveForFutureUse != nil && *payload.SaveForFutureUse {
		if !slices.Contains([]string{constant.UnifiedPaymentMethodCard}, paymentMethodType) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToSavedFutureUse)
		}
	}
	if payload.ShowSavedPayment != nil && *payload.ShowSavedPayment {
		if payload.Mode != constant.UnifiedPaymentModeRedirect || !slices.Contains([]string{constant.UnifiedPaymentMethodCard}, paymentMethodType) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToShowSavedPayment)
		}
	}
	if payload.ExpirationMode != "" {
		if !slices.Contains([]string{constant.UnifiedPaymentMethodCard, constant.UnifiedPaymentMethodEWallet}, paymentMethodType) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToSetExpirationMode)
		}
	}

	// Validate recurring payment feature usage
	if payload.RecurringID == "" {
		// If the recurring ID is not provided, the transaction is not considered a recurring payment
		// and the validation does not apply; proceed to the next process.

	} else if payload.InitiateFirstAuthorization && (payload.PaymentMethod == nil || payload.PaymentMethod.Type != constant.UnifiedPaymentMethodCard) {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("%s", "First authorization for recurring payments is only supported for CARD payment methods"))

	} else if !payload.InitiateFirstAuthorization && payload.PaymentMethod != nil {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("%s", "Subsequent recurring payments are not allowed to provide a payment method"))

	} else if payload.Mode != constant.UnifiedPaymentModeAPI {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("%s", "Recurring payments can only be created using API mode"))

	} else if !payload.InitiateFirstAuthorization && !payload.AutoConfirm {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("%s", "For subsequent recurring payments, the autoConfirm value must be TRUE"))

	} else if payload.CustomerID != "" || payload.CustomerInformation != nil {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("%s", "Customer ID or Customer Object are not required for recurring payments"))
	}

	if payload.PaymentMethod != nil {
		if payload.PaymentType == constant.UnifiedPaymentTypeMultiple &&
			!slices.Contains([]string{constant.UnifiedPaymentMethodQris, constant.UnifiedPaymentMethodVA}, payload.PaymentMethod.Type) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedForThisMethod)
		}

		switch payload.PaymentMethod.Type {
		case constant.UnifiedPaymentMethodVA:
			if payload.PaymentMethodOptions.VirtualAccount == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for virtual account can not be empty"))
			}
			if !payload.IsStaticPayment() || (payload.IsStaticPayment() && payload.Amount.Value > 0) {
				if payload.Amount.Value == 0 {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
				}

				conf := c.config.UnifiedPaymentConfig.VirtualAccountConfig
				if conf != nil {
					if err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount); err != nil {
						return err
					}
				}
			}

		case constant.UnifiedPaymentMethodEWallet:
			if payload.PaymentMethodOptions.Ewallet == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for ewallet can not be empty"))
			}
			if payload.Amount.Value == 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
			}
			conf := c.config.UnifiedPaymentConfig.EwalletConfig
			if conf != nil {
				if err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount); err != nil {
					return err
				}
			}

			if !payload.ExpiryAt.IsZero() {
				switch strings.ToUpper(payload.PaymentMethodOptions.Ewallet.Channel) {
				case constant.UnifiedPaymentEWalletShopeePayAcquirer:
					if payload.ExpiryAt.Sub(time.Now().UTC()) > constant.EWalletShopeePayMaxExpiryTime {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrEWalletShopeePayExceedMaxExpiryTime)
					}
				case constant.UnifiedPaymentEWalletDanaAcquirer:
					if payload.ExpiryAt.Sub(time.Now().UTC()) > constant.EWalletDanaMaxExpiryTime {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrEWalletDanaExceedMaxExpiryTime)
					}
				}
			}

			if payload.ExpirationMode == "" {
				payload.ExpirationMode = constant.UnifiedPaymentExpirationModeLoose
			}

		case constant.UnifiedPaymentMethodCard:
			if util.ValueOfPtr(payload.PaymentMethodOptions.Card).ThreeDsMethod == constant.CardThreeDsMethodExternal {
				// Validate Mode, should use API mode for threeDsMethod = EXTERNAL
				if payload.Mode != constant.UnifiedPaymentModeAPI {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentModeMustBeAPIForThreeDsMethodExternal)
				}

				// Validate that threeDsMethod = EXTERNAL should not be allowed to use autoConfirm = TRUE
				if payload.AutoConfirm {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentAutoConfirmMustBeFalseForThreeDsMethodExternal)
				}
			}

			if payload.Mode == constant.UnifiedPaymentModeAPI && payload.AutoConfirm && payload.PaymentMethod.CardPaymentMethodDetail == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method card can not be empty"))
			}
			if payload.RecurringID == "" && payload.Amount.Value == 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
			}
			conf := c.config.UnifiedPaymentConfig.CardConfig
			if (payload.RecurringID == "" && !payload.IsAutoSplitCardPayment()) && conf != nil {
				err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount)
				if err != nil && !c.isCybersourceTestAmountAllowed(payload.MerchantID, payload.Amount.Value) {
					return err
				}
			}

			if payload.ExpirationMode == "" {
				payload.ExpirationMode = constant.UnifiedPaymentExpirationModeLoose
			}

		case constant.UnifiedPaymentMethodQris:
			if !payload.IsStaticPayment() {
				if payload.Amount.Value == 0 {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
				}

				conf := c.config.UnifiedPaymentConfig.QrConfig
				if conf != nil {
					if err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount); err != nil {
						return err
					}
				}
			}

			if payload.IsStaticPayment() && payload.Amount.Value > 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentStaticQrAmountMustBeZero)
			}
		}

		// Validate that threeDsMethod EXTERNAL only available for CARD method
		if payload.PaymentMethod.Type != constant.UnifiedPaymentMethodCard &&
			util.ValueOfPtr(payload.PaymentMethodOptions.Card).ThreeDsMethod == constant.CardThreeDsMethodExternal {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentThreeDsMethodExternalOnlySupportForCard)
		}
	}

	if payload.SplitRoutingConfigurations != nil && len(*payload.SplitRoutingConfigurations) > 0 {
		requestCurrency := payload.Amount.Currency
		requestAmount := payload.Amount.Value

		totalSplitRoutingAmount := 0.0
		for _, value := range *payload.SplitRoutingConfigurations {
			if requestCurrency != value.Currency {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("currency is not match"))
			}

			calculationAmount := value.FixedAmount
			if value.Type == constant.SplitRoutingPaymentTypePercentage {
				calculationAmount = (value.PercentageAmount / 100) * requestAmount
			}

			totalSplitRoutingAmount += calculationAmount
		}

		if requestAmount < totalSplitRoutingAmount {
			return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("total split and routing amount must be not greater than payment amount"))
		}
	}

	ffContext := ffcontext.NewEvaluationContext(payload.MerchantID)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, payload.MerchantID)
	allowSpecialChars, _ := ffclient.BoolVariation(constant.FeatureFlagKeyUnifiedPaymentClientReferenceSpecialCharsWhitelistedMerchant, ffContext, false)
	if !allowSpecialChars && !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(payload.ClientReferenceID) {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrClientReferenceIDMustBeInAlphanumericFormat)
	}

	// validate metadata based on the characters only
	// the error cannot be reproduced due to come from marshaled payload req
	bytes, _ := json.Marshal(payload.Metadata)
	if utf8.RuneCount(bytes) > constant.UnifiedPaymentMaxMetadataLength {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentMetadataSizeLimitExceeded)
	}

	isRedirectMode := payload.Mode == constant.UnifiedPaymentModeRedirect
	isRedirectUrlEmpty := payload.RedirectUrl.SuccessReturnUrl == "" ||
		payload.RedirectUrl.FailureReturnUrl == "" ||
		payload.RedirectUrl.ExpirationReturnUrl == ""

	if isRedirectMode && payload.BypassStatusPage && isRedirectUrlEmpty {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage)
	}

	return nil
}

func (c *paymentController) simulateApiError(ctx context.Context, payload *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) (err error) {
	if c.config.Environment == constant.EnvironmentProduction {
		return nil
	}

	switch payload.Amount.Value {
	case 100033:
		err = pkgErrors.New(response.HttpErrRequest, constant.ErrMessageForApiErrorSimulation)
	case 100044:
		err = pkgErrors.New(response.HttpErrForbidden, constant.ErrMessageForApiErrorSimulation)
	case 100055:
		err = pkgErrors.New(response.HttpErrNotFound, constant.ErrMessageForApiErrorSimulation)
	case 100066:
		err = pkgErrors.New(response.HttpErrDupCheck, constant.ErrMessageForApiErrorSimulation)
	case 100088:
		err = pkgErrors.New(response.HttpErrRequestLimitExceeded, constant.ErrMessageForApiErrorSimulation)
	case 100099:
		err = pkgErrors.New(response.HttpErrInternal, constant.ErrMessageForApiErrorSimulation)
	case 110000:
		err = pkgErrors.New(response.HttpErrBadGateway, constant.ErrMessageForApiErrorSimulation)
	case 220000:
		err = pkgErrors.New(response.HttpErrServiceUnavailable, constant.ErrMessageForApiErrorSimulation)
	case 330000:
		err = pkgErrors.New(response.HttpErrRequestTimeout, constant.ErrMessageForApiErrorSimulation)
	}

	if err != nil {
		c.logger.Info(ctx, "[v2UnifiedPayment-Create] Entering api error simulation", logger.String("errorMessage", err.Error()))
		return err
	}

	return nil
}

// ValidateCustomerPayload validates the customer information in a unified payment session request.
// It handles three scenarios:
// 1. If neither CustomerID nor CustomerInformation is provided, validation passes.
// 2. If both CustomerID and CustomerInformation are provided, returns an error for conflict.
// 3. If only CustomerInformation is provided, creates a new customer and sets the CustomerID.
// 4. If only CustomerID is provided, verifies the customer exists for the given merchant.
func (c *paymentController) ValidateCustomerPayload(ctx context.Context, payload *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/internalController/unifiedPayment/ValidateCustomerPayload")
	defer segment.End()

	if payload.CustomerID == "" && payload.CustomerInformation == nil {
		return nil
	}

	if payload.CustomerID != "" && payload.CustomerInformation != nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrCustomerInformationConflict)
	}

	if payload.CustomerInformation != nil {
		createCustomerPayload := customerModel.CreateUnifiedPaymentCustomerRequest{
			MerchantID: payload.MerchantID,
			FirstName:  payload.CustomerInformation.GivenName,
			LastName:   payload.CustomerInformation.GetSurname(),
			Email:      payload.CustomerInformation.Email,
			Metadata: map[string]interface{}{
				"refundPreference": payload.CustomerInformation.RefundPreference,
			},
		}

		if payload.CustomerInformation.PhoneNumber != nil {
			createCustomerPayload.PhoneNumber = payload.CustomerInformation.PhoneNumber.Number
			createCustomerPayload.PhoneCountryCode = payload.CustomerInformation.PhoneNumber.CountryCode
		}

		customer, err := c.customerSvc.CreateUnfiedPaymentCustomer(ctx, createCustomerPayload)
		if err != nil {
			return err
		}

		payload.CustomerID = customer.UUID
		return nil
	}

	cust, err := c.customerSvc.GetCustomerById(ctx, payload.CustomerID, payload.MerchantID)
	if err != nil {
		return err
	}

	if cust == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrCustomerNotFound)
	}

	return nil
}
