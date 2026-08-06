package orchestratorController

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDetailById	godoc
// @Summary			Endpoint to display transaction details in the transaction history list
// @Description		Endpoint to display transaction details in the transaction history list
// @ID				transaction-histories-detail
// @Tags			Transaction History - Detail
// @Accept			json
// @Produce			json
// @Param 			transaction_id	path	string true "Transaction Id (uuid)"
// @Success			200  	{object}	response.ApiResponse{data=orchestrator_model.TransactionHistoryDetailResp}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/transaction-histories/details/{transaction_id} [get]
// @Security 		Bearer
func (c *OrchestratorController) GetDetailById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/orchestrator/GetDetailById")
	defer segment.End()

	id := chi.URLParam(r, "transaction_id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid transaction id format")))
		return
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	resp, err := c.orchestratorSvc.GetDetailById(ctx, user.MerchantId, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
