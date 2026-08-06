package recurring

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetRecurringByID		godoc
// @Summary		Get recurring contract by ID Endpoint
// @Description	Get recurring contract details by ID
// @ID			api-recurrings-get-by-id
// @Tags		API - Recurring Contracts
// @Accept		json
// @Produce		json
// @Param		uuid	path		string	true	"Recurring Contract UUID"
// @Success		200  	{object}	response.ApiResponse{data=recurringContractModel.GetRecurringByIDResponse}
// @Failure		400  	{object}	response.ApiResponse
// @Failure		401  	{object}	response.ApiResponse
// @Failure		404  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/recurrings/{uuid} [get]
// @Security	Bearer
func (c *RecurringContractController) GetRecurringByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/recurring/GetRecurringByID")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	recurringID := chi.URLParam(r, "uuid")
	if recurringID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err := uuid.Validate(recurringID); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	request := recurringContractModel.GetRecurringByIDRequest{
		RecurringID: recurringID,
		MerchantID:  user.MerchantId,
	}

	result, err := c.recurringContractService.GetRecurringByID(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
