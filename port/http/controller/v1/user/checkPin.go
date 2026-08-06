package user

import (
	"encoding/json"
	e "errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CheckCurrentPin	godoc
// @Summary			Check PIN endpoint
// @Description		Check PIN endpoint
// @ID				api-user-check-pin
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param			Request	body		user.CheckPinRequest true "JSON Body for Check Password"
// @Success			200  	{object}	response.ApiResponse{data=string}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/users/pin/check [post]
// @Security		Bearer
func (c *UserController) CheckCurrentPin(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/CheckCurrentPin")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.CheckPinRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.userSvc.CheckCurrentPin(ctx, user.UUID, payload.Pin); err != nil {
		if e.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
			// publish activity, do nothing on error
			_ = c.rabbitMqExt.PublishActivity(
				ctx,
				&user.MerchantId,
				&user.UUID,
				constant.TagAccount,
				constant.ErrFailedCheckPassword.Error(),
				map[string]string{
					"email": user.Email,
				},
			)
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, "OK")
}
