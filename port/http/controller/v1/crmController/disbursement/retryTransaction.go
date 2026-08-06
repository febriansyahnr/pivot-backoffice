package crmDisbursementController

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RetryTransactionForDisbursement	godoc
// @Summary				Retry transaction for disbursement
// @Description			Retry transaction for disbursement. Use forceFailed=true to force NOT_FOUND (07) status to FAILED in SNAP Core
// @ID					crm-retry-transaction-for-disbursement
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id			path	string	true	"ID disbursement"
// @Param 				forceFailed	query	bool	false	"Force NOT_FOUND status to FAILED (default: false)"
// @Param 				Request		body	disbursementModel.RetryTransaction true "Form Body for Send"
// @Success				200  		{object}	response.Response
// @Failure				500  		{object}	response.Response
// @Router				/crm/v1/disbursements/{id}/retry-transaction [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) RetryTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/disbursement/RetryTransaction")
	defer segment.End()

	disbursementID := chi.URLParam(r, "id")
	if err := uuid.Validate(disbursementID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	request := &disbursementModel.RetryTransaction{
		DisbursementID: disbursementID,
		ForceFailed:    false, // Default to false
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	if err := h.disbursementSvc.RetryDueToInsufficientEscrowFund(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{"id": disbursementID})
}
