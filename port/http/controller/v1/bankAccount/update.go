package bankAccount

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"net/http"
)

// Update		godoc
// @Summary		Endpoint for update bank account for withdrawal process
// @Description	Endpoint for update bank account for withdrawal process
// @ID			bank-account-update
// @Tags		API - Bank Account
// @Accept		json
// @Produce		json
// @Param 		id	path		string true "Bank Account ID to update role"
// @Param		Request	body		bankAccount.BankAccountRequest true "JSON Body for update bank account request"
// @Success		200  	{object}	response.ApiResponse{data=bankAccount.BankAccountResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/bank-accounts/:id [post]
// @Security 	Bearer
func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/bankAccount/Update")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	var payload bankAccount.UpdateBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.StructCtx(ctx, &payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	payload.UUID = id
	if resp, err := h.service.Update(ctx, &payload); err != nil {
		response.SendApiResponseError(ctx, w, err)
	} else {
		response.SendApiResponseOK(w, resp)
	}
}
