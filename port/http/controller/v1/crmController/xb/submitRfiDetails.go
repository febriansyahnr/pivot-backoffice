package crmXbController

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// SubmitRfiDetails	godoc
// @Summary				Submit Rfi Details for XB Payout
// @Description			Submit Rfi Details for XB Payout
// @ID					submit-rfi-details-for-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id	path		string true "XB Payout ID"
// @Param 				Request	body		xbModel.SubmitRfiDetailsRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=xbModel.SubmitRfiDetailsResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/payout/{id}/submit-rfi [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) SubmitRfiDetails(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		request *xbModel.SubmitRfiDetailsRequest
		ctx     = r.Context()
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/xb/SubmitRfiDetails")
	defer segment.End()

	payoutId := chi.URLParam(r, "id")
	if err := uuid.Validate(payoutId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	request = &xbModel.SubmitRfiDetailsRequest{
		PayoutId:   payoutId,
		MerchantId: r.FormValue("merchantId"),
		DocumentId: r.FormValue("documentId"),
		Comment:    r.FormValue("comment"),
		Value:      r.FormValue("value"),
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	if _, request.Document, err = r.FormFile("document"); err != nil && err != http.ErrMissingFile {
		response.SendGeneralResponseError(w, err)
		return
	}

	// Value and document cannot be exist at the same time
	if request.Value != "" && request.Document != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("value and document cannot be exist at the same time")))
		return
	}

	// Either value or document must be exist
	if request.Value == "" && request.Document == nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("either value or document must be exist")))
		return
	}

	rfiDetails, err := h.XbPayoutSvc.SubmitRfiDetails(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, rfiDetails)
}
