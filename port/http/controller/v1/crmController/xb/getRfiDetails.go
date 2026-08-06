package crmXbController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetRfiDetails	godoc
// @Summary				Get Rfi Details for XB Payout
// @Description			Get Rfi Details for XB Payout
// @ID					get-rfi-details-for-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id	path		string true "XB Payout ID"
// @Param 				Request	body		xbModel.GetRfiDetailsRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=xbModel.GetRfiDetailsResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/payout/{id}/get-rfi [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetRfiDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/xb/GetRfiDetails")
	defer segment.End()

	payoutId := chi.URLParam(r, "id")
	if err := uuid.Validate(payoutId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	request := &xbModel.GetRfiDetailsRequest{
		PayoutId: payoutId,
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	rfiDetails, err := h.XbPayoutSvc.GetRfiDetails(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, rfiDetails)
}
