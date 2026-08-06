package v2InternalUnifiedPaymentController

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *paymentController) Confirm(w http.ResponseWriter, r *http.Request) {
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

	id := chi.URLParam(r, "uuid")
	if err = uuid.Validate(id); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
		PaymentSessionID: id,
		MerchantID:       merchantID,
		ParentMerchantID: merchantAuth.MerchantId,
	}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err = c.validate.Struct(request); err != nil {
		c.logger.Warn(ctx, "confirm session validation error", logger.Error(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	payment, err := c.unifiedPaymentSvc.GetSessionDetail(ctx, &unifiedPaymentModel.GetUnifiedPaymentSessionRequest{
		PaymentSessionID: request.PaymentSessionID,
		MerchantID:       request.MerchantID,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	request.Amount = payment.Amount
	request.AutoSplitPayment = payment.AutoSplitPayment

	if request.PaymentMethod == nil {
		request.PaymentMethod = payment.PaymentMethod
	}
	if request.PaymentType == "" {
		request.PaymentType = payment.PaymentType
	}
	if payment.RecurringID != "" {
		request.RecurringID = payment.RecurringID
		request.InitiateFirstAuthorization = payment.InitiateFirstAuthorization
		request.FirstAuthorizationMethod = payment.FirstAuthorizationMethod
		request.FirstAuthorizationOrderID = payment.FirstAuthorizationOrderID
		request.RecurringBillingCycle = payment.RecurringBillingCycle
	}

	if errValidate := c.validateConfirmPayload(request); errValidate != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errValidate)
		return
	}

	if resp, err = c.unifiedPaymentSvc.ConfirmSession(ctx, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *paymentController) validateConfirmPayload(payload *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) error {
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
					if err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount); err != nil {
						return err
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
				if err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount); err != nil {
					return err
				}
			}

		case constant.UnifiedPaymentMethodCard:
			if payload.Mode == constant.UnifiedPaymentModeAPI && payload.PaymentMethod.CardPaymentMethodDetail == nil {
				return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method card can not be empty"))
			}
			if payload.RecurringID == "" && payload.Amount.Value == 0 {
				return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired)
			}
			conf := c.config.UnifiedPaymentConfig.CardConfig
			if payload.AutoSplitPayment == nil && conf != nil {
				err := c.validateAmountRange(payload.Amount.Value, conf.MinAmount, conf.MaxAmount)
				if err != nil && !c.isCybersourceTestAmountAllowed(payload.MerchantID, payload.Amount.Value) {
					return err
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
