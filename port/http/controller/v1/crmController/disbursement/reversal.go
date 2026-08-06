package crmDisbursementController

import (
	"encoding/json"
	"errors"
	"net/http"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// Reversal		godoc
// @Summary		Reversal of disbursement transaction via internal dashboard
// @Description	Reversal of disbursement transaction via internal dashboard
// @ID			crm-disbursement-transaction-reversal
// @Tags		API - CRM
// @Accept		json
// @Produce		json
// @Param 		id		path	string	true	"Disbursement Id"
// @Param 		Request	body	disbursementModel.ReversalTransactionReq true "Form Body for reversal disbursement transaction"
// @Success		200  	{object}	response.Response{data=disbursementModel.ReversalTransactionResp}
// @Failure		500  	{object}	response.Response
// @Router		/crm/v1/disbursements/{id}/reversal [post]
// @Header      all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) Reversal(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/disbursement/Reversal")
	defer segment.End()

	request := &disbursementModel.ReversalTransactionReq{
		DisbursementId: r.PathValue("id"),
	}

	if err := uuid.Validate(request.DisbursementId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid disbursement id")))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.disbursementSvc.Reversal(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
