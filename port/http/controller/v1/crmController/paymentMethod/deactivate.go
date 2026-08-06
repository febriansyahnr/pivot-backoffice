package crmPaymentMethodController

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// DeactivatePaymentMethod	godoc
// @Summary					Deactivate payment method by ID
// @Description				Deactivate payment method by ID
// @ID						crm-deactivate-payment-method-by-id
// @Tags					API - CRM
// @Accept					mpfd
// @Produce					mpfd
// @Param 					id					path		string true "ID merchant"
// @Param 					paymentMethodId		path		string true "paymentMethodId for deactivate"
// @Success					200  		{object}	response.Response
// @Failure					500  		{object}	response.Response
// @Router					/crm/v1/merchants/{id}/payment-methods/{paymentMethodId}/deactivate [patch]
// @Header       			all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) DeactivatePaymentMethodMerchant(w http.ResponseWriter, r *http.Request) {
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

	// find payment method by ID.
	paymentMethod, err := h.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, paymentMethodId, merchantID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	if !paymentMethod.IsActive {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrPaymentMethodAlreadyInactive))
		return
	}

	// Call deactivate service
	if errDeactivate := h.paymentMethodSvc.Deactivate(ctx, paymentMethod); errDeactivate != nil {
		response.SendGeneralResponseError(w, errDeactivate)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"merchantId":      merchantID,
		"paymentMethodId": paymentMethodId,
	})
}
