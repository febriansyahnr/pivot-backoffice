package disbursementController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RetryBulk		godoc
// @Summary			Cancel disbursement
// @Description		Cancel disbursement
// @ID				api-cancel-disbursements
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Param			Request	body		disbursementModel.RetryBulkRequest true "JSON Body for retry bulk disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/disbursements/cancel [post]
// @Security		Bearer
func (c *Controller) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/Cancel")
	defer segment.End()

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err := constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	var payload disbursementModel.CancelPayoutRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Set merchant ID
	payload.MerchantID = user.MerchantId

	cancelledPayouts, err := c.disbursementSvc.Cancel(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, disbursementModel.CancelDisbursementResponse{
		Total:        len(cancelledPayouts),
		CancelledIds: cancelledPayouts,
	})
}
