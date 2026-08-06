package v1InternalUnifiedPaymentController

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
)

func (c *paymentController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/Create")
	defer segment.End()

	var (
		err error
		now = time.Now()
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

	payload := paymentModel.CreateUnifiedPaymentRequest{
		MerchantID: merchantID,
		CreatedBy:  merchantAuth.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err := c.validate.Struct(payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer monitor.WriteAndSend(
		ctx, "internal-controller-payment-v2-create", now, ww, err, nil,
	)

	if errValidate := c.validatePayload(&payload); errValidate != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errValidate)
		return
	}

	// Call service to create unified payment v2
	if c.isPaymentMigrationV1ToV2Enabled(ctx, merchantID) {
		v2Request := payload.ToCreateUnifiedPaymentSessionRequest()

		if errValidate := c.ValidateCustomerPayload(ctx, v2Request); errValidate != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, errValidate)
			return
		}

		resp, err := c.unifiedPaymentSvc.CreateSession(ctx, v2Request)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}

		response.SendApiResponseOK(w, paymentModel.MapUnifiedPaymentV2ToV1Response(resp))
		return
	}

	// TODO: Remove later once new unified payment service is stable
	resp, err := c.paymentSvc.CreateUnifiedPayment(ctx, &payload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
	}

	response.SendApiResponseOK(w, resp)
}

func (c *paymentController) validatePayload(payload *paymentModel.CreateUnifiedPaymentRequest) error {
	if payload.ExpiryAt.Before(time.Now().UTC()) {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("expiry is not allowed to be less than current time"))
	}

	if payload.PaymentMethod == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT && payload.PaymentMethodOptions.VirtualAccount == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method option virtual account is required"))
	}

	if payload.PaymentMethod == paymentConstant.PAYMENT_METHOD_CREDIT_CARD && payload.PaymentMethodOptions.Card == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method option card is required"))
	}

	if payload.SplitRoutingConfigurations != nil && len(*payload.SplitRoutingConfigurations) > 0 {
		requestCurrency := payload.Amount.Currency
		requestAmount, _ := payload.Amount.Value.Float64()

		totalSplitRoutingAmount := 0.0
		for _, value := range *payload.SplitRoutingConfigurations {
			if requestCurrency != value.Currency {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("currency is not match"))
			}

			calculationAmount := value.FixedAmount
			if value.Type == constant.SplitRoutingPaymentTypePercentage {
				calculationAmount = (value.PercentageAmount / 100) * requestAmount
			}

			totalSplitRoutingAmount += decimal.NewFromFloat(calculationAmount).Round(2).InexactFloat64()
		}

		if requestAmount < totalSplitRoutingAmount {
			return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("total split and routing amount must be not greater than payment amount"))
		}
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
