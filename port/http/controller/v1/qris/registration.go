package qris

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Registration		godoc
// @Summary			This endpoint is used for QRIS registration.
// @Description		This endpoint is used for QRIS registration where the merchant has completed KYC/KYB in full.
// @ID				qris-registration
// @Tags			API - QRIS Registration
// @Accept			json
// @Produce			json
// @Param			Request	body		qris.RegistrationReq	true	"JSON Body for QRIS registration"
// @Success			200  	{object}	response.Response{data=qris.RegistrationResp}
// @Failure			500  	{object}	response.Response
// @Router			/crm/v1/qr/registrations [post]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) Registration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/qris/Registration")
	defer segment.End()

	var request qris.RegistrationReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if id, err := h.service.Registration(ctx, &request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, qris.RegistrationResp{Id: id})
	}
}
