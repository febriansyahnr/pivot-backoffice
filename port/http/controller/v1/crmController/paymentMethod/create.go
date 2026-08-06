package crmPaymentMethodController

import (
	"encoding/json"
	"net/http"

	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ActivatePaymentMethod	godoc
// @Summary					Create new Payment Method
// @Description				Enable ops team to define new payment method
// @ID						crm-activate-all-payment-method
// @Tags					API - CRM
// @Success					200  		{object}	response.Response
// @Failure					500  		{object}	response.Response
// @Router					/crm/v1/payment-methods [patch]
// @Header       			all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/paymentMethod/Create")
	defer segment.End()

	var payload = &paymentMethodModel.CreatePaymentMethodRequest{}

	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	err = h.validator.Struct(payload)
	if err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	err = h.paymentMethodSvc.Create(ctx, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"created": true,
	})
}
