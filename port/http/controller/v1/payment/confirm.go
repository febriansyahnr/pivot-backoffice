package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PaymentController) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/ConfirmPayment")
	defer segment.End()

	var err error

	ctx = context.WithValue(ctx, constant.CtxFeatureName, constant.FeatureConfirmPaymentCheckoutUI)

	// Get payment ID from context (set by PaymentTokenMiddleware)
	paymentID, ok := ctx.Value(constant.CtxPaymentID).(string)
	if !ok || paymentID == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	// Parse request body
	request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	// Set payment session ID from token
	request.PaymentSessionID = paymentID

	// Validate request
	if err = c.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Get payment details to populate missing fields (using GetPaymentDetailForPaymentUI since we don't have merchant ID yet)
	paymentDetail, err := c.paymentService.GetPaymentDetailForPaymentUI(ctx, paymentID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Set merchant ID from payment details
	request.MerchantID = paymentDetail.MerchantID

	// Now get full payment details with merchant ID
	payment, err := c.unifiedPaymentService.GetSessionDetail(ctx, &unifiedPaymentModel.GetUnifiedPaymentSessionRequest{
		PaymentSessionID: paymentID,
		MerchantID:       paymentDetail.MerchantID,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Set defaults from existing payment if not provided
	if request.PaymentType == "" {
		// Determine payment type from existing payment metadata
		// For now, assume SINGLE unless metadata indicates otherwise
		request.PaymentType = constant.UnifiedPaymentTypeSingle

		// Check if this is a MULTIPLE type payment from payment type
		if payment.PaymentType == constant.UnifiedPaymentTypeMultiple {
			request.PaymentType = constant.UnifiedPaymentTypeMultiple
		}
	}

	// Set amount from existing payment
	request.Amount = payment.Amount

	// Validate confirm payload
	if errValidate := c.validateConfirmPayload(request); errValidate != nil {
		response.SendApiResponseError(ctx, w, errValidate)
		return
	}

	// Use Unified Payment V2 service directly
	if c.unifiedPaymentService == nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrInternal, errors.New("unified payment service not available")))
		return
	}

	var unifiedResp *unifiedPaymentModel.UnifiedPaymentSessionResponse
	if unifiedResp, err = c.unifiedPaymentService.ConfirmSession(ctx, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, unifiedResp)
}

// validateConfirmPayload validates the confirm request payload
func (c *PaymentController) validateConfirmPayload(payload *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) error {
	if payload.PaymentMethod != nil {
		if payload.PaymentType == constant.UnifiedPaymentTypeMultiple &&
			!slices.Contains([]string{constant.UnifiedPaymentMethodQris, constant.UnifiedPaymentMethodVA}, payload.PaymentMethod.Type) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedForThisMethod)
		}

		switch payload.PaymentMethod.Type {
		case constant.UnifiedPaymentMethodVA:
			if payload.PaymentMethodOptions == nil || payload.PaymentMethodOptions.VirtualAccount == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for virtual account can not be empty"))
			}
			if !payload.IsStaticPayment() || (payload.IsStaticPayment() && payload.Amount.Value > 0) {
				if payload.Amount.Value == 0 {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
				}
				conf := c.config.UnifiedPaymentConfig.VirtualAccountConfig
				if conf != nil {
					if conf.MinAmount != nil && *conf.MinAmount > payload.Amount.Value {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount)
					}
					if conf.MaxAmount != nil && *conf.MaxAmount < payload.Amount.Value {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount)
					}
				}
			}

		case constant.UnifiedPaymentMethodEWallet:
			if payload.PaymentMethodOptions == nil || payload.PaymentMethodOptions.Ewallet == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for ewallet can not be empty"))
			}
			if payload.Amount.Value == 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
			}
			conf := c.config.UnifiedPaymentConfig.EwalletConfig
			if conf != nil {
				if conf.MinAmount != nil && *conf.MinAmount > payload.Amount.Value {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount)
				}
				if conf.MaxAmount != nil && *conf.MaxAmount < payload.Amount.Value {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount)
				}
			}

		case constant.UnifiedPaymentMethodCard:
			if payload.PaymentMethod.CardPaymentMethodDetail == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method card can not be empty"))
			}
			if payload.Amount.Value == 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
			}
			conf := c.config.UnifiedPaymentConfig.CardConfig
			if conf != nil {
				if conf.MinAmount != nil && *conf.MinAmount > payload.Amount.Value {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount)
				}
				if conf.MaxAmount != nil && *conf.MaxAmount < payload.Amount.Value {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount)
				}
			}

		case constant.UnifiedPaymentMethodQris:
			if !payload.IsStaticPayment() {
				if payload.Amount.Value == 0 {
					return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
				}

				conf := c.config.UnifiedPaymentConfig.QrConfig
				if conf != nil {
					if conf.MinAmount != nil && *conf.MinAmount > payload.Amount.Value {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount)
					}
					if conf.MaxAmount != nil && *conf.MaxAmount < payload.Amount.Value {
						return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount)
					}
				}
			}

			if payload.IsStaticPayment() && payload.Amount.Value > 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentStaticQrAmountMustBeZero)
			}
		}
	}

	return nil
}
