package crmDisbursementController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// InquiryTransactionForDisbursement	godoc
// @Summary				Inquiry transaction for disbursement
// @Description			Inquiry transaction for disbursement
// @ID					crm-inquiry-transaction-for-disbursement
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id	path		string true "ID disbursement"
// @Param 				Request	body		disbursementModel.InquiryTransaction true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=disbursementModel.DisbursementWithTransactionResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/disbursements/{id}/inquiry-transaction [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) InquiryTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/disbursement/InquiryTransaction")
	defer segment.End()

	disbursementID := chi.URLParam(r, "id")
	if err := uuid.Validate(disbursementID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	request := &disbursementModel.InquiryTransaction{
		DisbursementID: disbursementID,
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	disbursement, err := h.disbursementSvc.InquiryTransaction(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, disbursement.DisbursementWithTransactionToResponse())
}
