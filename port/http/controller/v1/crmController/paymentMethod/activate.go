package crmPaymentMethodController

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ActivatePaymentMethod	godoc
// @Summary					Activate payment method by ID
// @Description				Activate payment method by ID
// @ID						crm-activate-payment-method-by-id
// @Tags					API - CRM
// @Accept					mpfd
// @Produce					mpfd
// @Param 					id					path		string true "ID merchant"
// @Param 					paymentMethodId		path		string true "paymentMethodId for activate"
// @Success					200  		{object}	response.Response
// @Failure					500  		{object}	response.Response
// @Router					/crm/v1/merchants/{id}/payment-methods/{paymentMethodId}/activate [patch]
// @Header       			all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) ActivatePaymentMethodMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/paymentMethod/ActivationPaymentMethodMerchant")
	defer segment.End()

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	paymentMethodId := chi.URLParam(r, "paymentMethodId")
	if err := uuid.Validate(paymentMethodId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("paymentMethodId is required")))
		return
	}

	request := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{UUID: paymentMethodId},
		MerchantID:    merchantID,
	}
	if errActivate := h.paymentMethodSvc.Activate(ctx, request); errActivate != nil {
		response.SendGeneralResponseError(w, errActivate)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"merchantId":      merchantID,
		"paymentMethodId": paymentMethodId,
	})
}

// ActivatePaymentMethod	godoc
// @Summary					Activate all payment method for merchant by ID
// @Description				allow merchant to use all payment method for ease of use
// @ID						crm-activate-all-payment-method
// @Tags					API - CRM
// @Accept					mpfd
// @Produce					mpfd
// @Param 					id			path		string true "Merchant ID"
// @Success					200  		{object}	response.Response
// @Failure					500  		{object}	response.Response
// @Router					/crm/v1/merchants/{id}/payment-methods/activate [patch]
// @Header       			all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) ActivateAllPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/paymentMethod/ActivateAllPaymentMethod")
	defer segment.End()

	if h.config.Environment != constant.EnvironmentStaging {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess))
		return
	}

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	// find payment method by ID.
	merchant, err := h.merchantSvc.FindMerchantByID(ctx, merchantID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	err = h.merchantSvc.EnableAllPaymentMethod(ctx, merchant)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"merchantId":      merchantID,
		"paymentMethodId": "all",
	})
}
