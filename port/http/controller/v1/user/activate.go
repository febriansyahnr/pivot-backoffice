package user

import (
	"context"
	"errors"

	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Activate		godoc
// @Summary		Activate user endpoint
// @Description	Activate user endpoint
// @ID			user-activation
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.ActivateUserRequest true "JSON Body for Activate User"
// @Success		200  	{object}	response.ApiResponse{data=user.UserLoggedInResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/activate [patch]
// @Security 	Bearer
func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/Activate")
	defer segment.End()

	tokenFromHeader := r.Header.Get("X-Invitation-Token")
	if tokenFromHeader == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	request := userModel.ActivateUserRequest{
		Token: tokenFromHeader,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	// validate activation token first
	if result, err := c.userSvc.ValidateInvitationToken(ctx, request.Token); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if result.Email != request.Email {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, errors.New("email mismatch")))
		return
	}

	// Add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserAgentKey, r.UserAgent())

	// activate user
	resp, err := c.userSvc.ActivateUser(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
