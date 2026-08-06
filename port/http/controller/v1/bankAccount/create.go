package bankAccount

import (
	"encoding/json"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"net/http"
)

// Create		godoc
// @Summary		Endpoint for create bank account for withdrawal process
// @Description	Endpoint for create bank account for withdrawal process
// @ID			bank-account-create
// @Tags		API - Bank Account
// @Accept		json
// @Produce		json
// @Param		Request	body		bankAccount.BankAccountRequest true "JSON Body for create bank account request"
// @Success		200  	{object}	response.ApiResponse{data=bankAccount.CreateBankAccountResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/bank-accounts [post]
// @Security 	Bearer
func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/bankAccount/Create")
	defer segment.End()

	var payload bankAccount.CreateBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.StructCtx(ctx, &payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.Create(ctx, &payload); err != nil {
		response.SendApiResponseError(ctx, w, err)
	} else {
		response.SendApiResponseOK(w, resp)
	}
}
