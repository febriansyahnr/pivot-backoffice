package crmfdscontroller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// TransactionEvaluation		godoc
// @Summary			FDS update transaction evaluation
// @Description		Use upate for transaction
// @ID				crm-update-transaction
// @Tags			CRM - FDS
// @Accept			json
// @Produce			json
// @Param			id		query		string					true	"id of account transactions"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/internal/v1/fds/update/:id [post]
// @Security		Bearer
func (c *CRMFdsController) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		payload fdscommon.CRMUpdateRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/fds/UpdateTransaction")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("account trx id is required")))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.fdsService.UpdateTransaction(ctx, id, payload.ToUpdateRequest())
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
