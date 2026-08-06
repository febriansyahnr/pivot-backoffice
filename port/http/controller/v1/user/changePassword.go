package user

import (
	"encoding/json"
	e "errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ChangePassword	godoc
// @Summary			Change Password endpoint
// @Description		Change Password endpoint
// @ID				api-user-change-password
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param			Request	body		user.ChangePasswordRequest true "JSON Body for Change Password"
// @Success			200  	{object}	response.ApiResponse{data=user.ChangePasswordResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/change-password [patch]
// @Security		Bearer
func (c *UserController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ChangePassword")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = fmt.Errorf("user not found")
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	var payload userModel.ChangePasswordRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	isSuccess, err := c.userSvc.ChangePassword(r.Context(), user.UUID, payload.OldPassword, payload.NewPassword)
	if err != nil {
		if e.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
			// publish activity, do nothing on error
			_ = c.rabbitMqExt.PublishActivity(
				ctx,
				&user.MerchantId,
				&user.UUID,
				constant.TagAccount,
				constant.ErrFailedToChangePassword.Error(),
				map[string]string{
					"email":        user.Email,
					"old_password": "********",
					"new_password": "********",
				},
			)
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	// publish activity, do nothing on error
	payload.OldPassword = "********"
	payload.NewPassword = "********"
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserChangePassword,
		payload,
	)

	response.SendApiResponseOK(w, isSuccess)
}
