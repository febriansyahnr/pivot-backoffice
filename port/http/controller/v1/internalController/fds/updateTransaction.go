package fds

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// TransactionEvaluation		godoc
// @Summary			FDS update transaction evaluation
// @Description		Use upate for transaction
// @ID				internal-transaction-evaluation
// @Tags			Internal - FDS
// @Accept			json
// @Produce			json
// @Param			id		query		string					true	"id of account transactions"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/internal/v1/fds/check/:id [post]
// @Security		Bearer
func (c *InternalFdsController) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/fds/UpdateTransaction")
	defer segment.End()

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("account trx id is required")))
		return
	}

	// eval using service
	resp, err := c.fdsSvc.UpdateTransaction(ctx, id, nil)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
