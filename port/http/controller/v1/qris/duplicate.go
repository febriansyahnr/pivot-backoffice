package qris

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// DuplicateRegistration	godoc
// @Summary			This endpoint is used to duplicate an existing QRIS registration to another merchant.
// @Description		This endpoint is used to duplicate a successful QRIS registration from one merchant to another.
// @ID				qris-registration-duplicate
// @Tags			API - QRIS Registration
// @Accept			json
// @Produce			json
// @Param			Request		body		qris.DuplicateRegistrationReq true	"JSON Body for duplicating registration"
// @Success			200  		{object}	response.Response{data=qris.DuplicateRegistrationResp}
// @Failure			500  		{object}	response.Response
// @Router			/crm/v1/qr/registrations/duplicate [post]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) DuplicateRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/qris/DuplicateRegistration")
	defer segment.End()

	// Environment check: Only allow this endpoint in non-production environments
	if h.config.Environment == constant.EnvironmentProduction {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess))
		return
	}

	var request qris.DuplicateRegistrationReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if id, err := h.service.DuplicateRegistration(ctx, &request); err != nil {
		response.SendGeneralResponseError(w, err)
	} else {
		response.SendGeneralResponseOK(w, qris.DuplicateRegistrationResp{Id: id})
	}
}
