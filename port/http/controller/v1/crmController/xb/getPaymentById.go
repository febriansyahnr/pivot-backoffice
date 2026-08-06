package crmXbController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPayoutByID	godoc
// @Summary				Get Payout Details for XB Payout
// @Description			Get Payout Details for XB Payout
// @ID					get-payout-details-for-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id		path		string true "XB Payout ID"
// @Param 				Request	body		xbModel.GetPayoutRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=xbModel.GetPayoutRequest}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/payout/{id} [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetPayoutByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/xb/GetPaymentByID")
	defer segment.End()

	payoutId := chi.URLParam(r, "id")
	if err := uuid.Validate(payoutId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	request := &xbModel.GetPayoutRequest{
		PayoutId: payoutId,
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	payout, err := h.XbPayoutSvc.GetPayoutById(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, payout)
}
