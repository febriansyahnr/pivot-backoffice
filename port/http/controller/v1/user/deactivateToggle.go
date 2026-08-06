package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// validateActivateDeactivate performs the common validation logic for both activate and deactivate.
func (c *UserController) validateActivateDeactivate(r *http.Request) (*userModel.ActivateDeactivateRequest, error) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/validateActivateDeactivate")
	defer segment.End()

	id := chi.URLParam(r, "user_id")
	if id == "" {
		return nil, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("user id is required"))
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		return nil, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
	}

	if user.UUID == id {
		return nil, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("unable to activate or deactivate one's own account"))
	}

	merchant, err := c.merchantSvc.FindMerchantByID(r.Context(), user.MerchantId)
	if err != nil {
		return nil, err
	}

	if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, fmt.Errorf("merchant with id %s not found", user.MerchantId))
	}

	existedUser, err := c.userSvc.FindUserByID(r.Context(), id)
	if err != nil {
		return nil, err
	}

	if existedUser.MerchantId != user.MerchantId {
		return nil, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
	}

	return &userModel.ActivateDeactivateRequest{
		UserID:         id,
		User:           existedUser,
		UserTokenClaim: user,
	}, nil
}

// ActivateUser		godoc
// @Summary			Activate user status
// @Description		Activate user status
// @ID				user-activate
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param 			user_id	path		string true "User ID to activate"
// @Success			200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/users/{user_id}/activate [put]
// @Security 		Bearer
func (c *UserController) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ActivateUser")
	defer segment.End()

	req, err := c.validateActivateDeactivate(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if req.User.Status == constant.UserStatusActive {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrAccountAlreadyActive))
		return
	}

	req.User.Email = strings.ToLower(req.User.Email)
	req.User.Status = constant.UserStatusActive
	req.User.DeactivatedAt = sql.NullTime{}
	req.User.UpdatedAt = time.Now().UTC()

	if err := c.userSvc.Update(ctx, req.User); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if errRemove := c.JWT.RemoveIterateTokenFromRedis(ctx, req.User.Email); errRemove != nil {
		response.SendApiResponseError(ctx, w, errRemove)
		return
	}

	response.SendApiResponseOK(w, req.User.ToResponse())
}

// DeactivateUser	godoc
// @Summary			Deactivate user status
// @Description		Deactivate user status
// @ID				user-deactivate
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param 			user_id	path		string true "User ID to deactivate"
// @Success			200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/users/{user_id}/deactivate [put]
// @Security 		Bearer
func (c *UserController) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/DeactivateUser")
	defer segment.End()

	req, err := c.validateActivateDeactivate(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if req.User.Status == constant.UserStatusInactive {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrAccountAlreadyInactive))
		return
	}

	req.User.Email = strings.ToLower(req.User.Email)
	req.User.Status = constant.UserStatusInactive
	req.User.DeactivatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	req.User.UpdatedAt = time.Now().UTC()

	if err := c.userSvc.Update(ctx, req.User); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if errRemove := c.JWT.RemoveIterateTokenFromRedis(ctx, req.User.Email); errRemove != nil {
		response.SendApiResponseError(ctx, w, errRemove)
		return
	}

	response.SendApiResponseOK(w, req.User.ToResponse())
}
