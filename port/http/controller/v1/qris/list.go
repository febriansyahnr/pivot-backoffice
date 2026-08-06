package qris

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// RegistrationList	godoc
// @Summary			This endpoint is used to retrieve the qris registration list.
// @Description		This endpoint is used to retrieve the qris registration list.
// @ID				qris-registration-list
// @Tags			API - QRIS Registration
// @Accept			json
// @Produce			json
// @Param			id			path		string	true	"Merchant Id"
// @Success			200  		{object}	response.Response{data=qris.RegistrationListResp}
// @Failure			500  		{object}	response.Response
// @Router			/crm/v1/qr/registrations/merchants/{id} [get]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) RegistrationList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/qris/GetRegistration")
	defer segment.End()

	merchantId := r.PathValue("id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	if resp, err := h.service.RegistrationList(ctx, merchantId); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
