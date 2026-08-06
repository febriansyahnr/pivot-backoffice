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
// @Summary			Retry bulk disbursement
// @Description		Retry bulk disbursement
// @ID				api-retry-bulk-disbursements
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Param			Request	body		disbursementModel.RetryBulkRequest true "JSON Body for retry bulk disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/disbursements/bulk/retry [post]
// @Security		Bearer
func (c *Controller) RetryBulk(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		ctx = r.Context()
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/disbursement/RetryBulk")
	defer func() {
		if err != nil {
			segment.End()
			return
		}

		segment.End()
	}()

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	var requestPayload disbursementModel.RetryBulkRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Set merchant ID
	requestPayload.MerchantID = user.MerchantId

	err = c.disbursementSvc.RetryBulk(r.Context(), &requestPayload)
	if err != nil {
		if resp, ok := err.(*disbursementModel.ApprovalResultErr); ok {
			response.SendApiResponseOK(w, resp)
			return
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
