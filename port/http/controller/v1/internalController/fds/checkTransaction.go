package fds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CheckTransaction		godoc
// @Summary			FDS transaction evaluation
// @Description		Use check for transaction
// @ID				internal-transaction-evaluation
// @Tags			Internal - FDS
// @Accept			json
// @Produce			json
// @Param			id		query		string					true	"id of account transactions"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/internal/v1/fds/check/:id [post]
// @Security		Bearer
func (c *InternalFdsController) CheckTransaction(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/fds/TransactionEvaluation")
	defer segment.End()

	// get id from url path
	id := chi.URLParam(r, "id")
	if errID := uuid.Validate(id); errID != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("account trx id is required")))
		return
	}

	var payload *fdscommon.CheckTransactionRequest
	// this allow to handle empty request body
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
			return
		}

		if err := c.validate.Struct(payload); err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
			return
		}
	}

	timeoutMs := c.GetTimeout()
	ctx, cancel := context.WithTimeout(ctx, (time.Duration(timeoutMs) * time.Millisecond))
	defer cancel()

	if c.config.Environment == constant.EnvironmentStaging {
		if cardNumber := r.Header.Get(constant.HeaderXSimulationCardNumber); cardNumber != "" {
			ctx = context.WithValue(ctx, constant.CtxTestCardNumber, cardNumber)
		}
	}

	// eval using service
	resp, err := c.fdsSvc.CheckTransaction(ctx, id, payload)
	if ctx.Err() != nil && (ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled) {
		c.logger.Error(ctx, constant.ErrFdsTimeout.Error(), logger.Error(err))
		// return response as is, since it is handled in the service
		response.SendOpenApiResponseOK(w, resp)
		return
	}

	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
