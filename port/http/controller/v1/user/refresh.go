package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Refresh		godoc
// @Summary		Refresh endpoint with email
// @Description	Refresh endpoint with email
// @ID			api-user-refresh-token
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserRefreshTokenRequest true "JSON Body for Refresh Token"
// @Success		200  	{object}	response.ApiResponse{data=user.UserLoggedInResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/logout [post]
func (c *UserController) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/Refresh")
	defer segment.End()

	// add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, r.Header.Get(constant.HeaderDeviceIdentifier))

	var payload userModel.UserRefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	user, signedToken, err := c.userSvc.Refresh(ctx, payload.Email, payload.RefreshToken)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if user == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, fmt.Errorf("user not found")))
		return
	}

	// fill response
	res := userModel.UserLoggedInResponse{
		UserInfo:     user.ToResponse(),
		AccessToken:  signedToken,
		RefreshToken: user.RefreshToken.String,
	}

	response.SendApiResponseOK(w, res)
}
