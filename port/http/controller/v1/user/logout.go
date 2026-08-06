package user

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Logout		godoc
// @Summary		Logout endpoint with email
// @Description	Logout endpoint with email
// @ID			api-user-logout
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserLogoutRequest true "JSON Body for Logout"
// @Success		200  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/logout [post]
// @Security	Bearer
func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/Logout")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.UserLogoutRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserAgentKey, r.UserAgent())

	// login
	err = c.userSvc.Logout(ctx, payload.Email)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Revoke all user's token
	if err = c.JWT.RemoveIterateTokenFromRedis(r.Context(), payload.Email); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserLogout,
		payload,
	)

	response.SendApiResponseOK(w, nil)
}
