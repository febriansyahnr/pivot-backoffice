package qris

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ReuploadDocument	godoc
// @Summary			This endpoint is used to re-upload documents for qris registration.
// @Description		This endpoint is used to re-upload documents for qris registration.
// @ID				qris-registration-reupload-documents
// @Tags			API - QRIS Registration
// @Accept			json
// @Produce			json
// @Param			Request		body		qris.ReuploadDocumentReq true	"JSON Body for reupload document"
// @Success			200  		{object}	response.Response{data=qris.ReuploadDocumentResp}
// @Failure			500  		{object}	response.Response
// @Router			/crm/v1/qr/registrations/documents [put]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) ReuploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/qris/ReuploadDocument")
	defer segment.End()

	var request qris.ReuploadDocumentReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.ReuploadDocument(ctx, &request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
