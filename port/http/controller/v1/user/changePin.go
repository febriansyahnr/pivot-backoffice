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

// ChangePin		godoc
// @Summary			Change PIN endpoint
// @Description		Change PIN endpoint
// @ID				api-user-change-pin
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param			Request	body		user.ChangePinRequest true "JSON Body for Change Password"
// @Success			200  	{object}	response.ApiResponse{data=user.ChangePinRequest}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/change-password [patch]
// @Security		Bearer
func (c *UserController) ChangePin(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ChangePin")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.ChangePinRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.userSvc.ChangePin(ctx, user.UUID, payload.Pin, payload.NewPin); err != nil {
		if e.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
			// publish activity, do nothing on error
			_ = c.rabbitMqExt.PublishActivity(
				ctx,
				&user.MerchantId,
				&user.UUID,
				constant.TagAccount,
				constant.ErrFailedToChangePINLimit.Error(),
				map[string]string{
					"email":   user.Email,
					"old_pin": "******",
					"new_pin": "******",
				},
			)
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	// publish activity, do nothing on error
	payload.Pin = "********"
	payload.NewPin = "********"
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserUpdatePIN,
		payload,
	)

	response.SendApiResponseOK(w, payload)
}
