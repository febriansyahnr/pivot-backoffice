package withdrawalCrmController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// RetryTransaction	godoc
// @Summary				Retry withdrawal bank transfer transaction
// @Description			Retry a failed/pending withdrawal bank transfer. Use forceFailed=true to force NOT_FOUND (07) status to FAILED in SNAP Core
// @ID					crm-retry-withdrawal-transaction
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param 				id			path	string	true	"Withdrawal ID"
// @Param 				forceFailed	query	bool	false	"Force NOT_FOUND status to FAILED (default: false)"
// @Param 				Request		body	withdrawal.RetryTransactionRequest true "Request body"
// @Success				200  		{object}	response.Response
// @Failure				500  		{object}	response.Response
// @Router				/crm/v1/withdrawals/{id}/retry-transaction [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) RetryTransaction(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/withdrawal/RetryTransaction")
	defer segment.End()

	withdrawalID := r.PathValue("id")
	if err := uuid.Validate(withdrawalID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	request := &withdrawal.RetryTransactionRequest{
		WithdrawalID: withdrawalID,
		ForceFailed:  false,
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	// Parse forceFailed query parameter (defaults to false if not provided)
	if forceFailedStr := r.URL.Query().Get("forceFailed"); forceFailedStr != "" {
		forceFailed, err := strconv.ParseBool(forceFailedStr)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest,
				fmt.Errorf("invalid forceFailed parameter: must be true or false")))
			return
		}
		request.ForceFailed = forceFailed
	}

	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.service.RetryTransaction(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{"id": withdrawalID})
}
