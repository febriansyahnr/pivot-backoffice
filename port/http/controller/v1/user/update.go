package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
)

// Update		godoc
// @Summary		Update user endpoint
// @Description	Update user endpoint
// @ID			user-update
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param 		user_id	path		string true "User ID to get user"
// @Param		Request	body		user.UserUpdateRequest true "JSON Body for Update User"
// @Success		200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/{user_id} [put]
// @Security 	Bearer
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/Update")
	defer segment.End()

	var (
		err error
	)

	targetUserID := chi.URLParam(r, "user_id")
	if targetUserID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("user id is required")))
		return
	}

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.UserUpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// check if merchant exists
	merchant, err := c.merchantSvc.FindMerchantByID(r.Context(), user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if merchant == nil {
		response.SendApiResponseError(ctx,
			w,
			errors.New(response.HttpErrNotFound, fmt.Errorf("merchant with id %s not found", user.MerchantId)))
		return
	}

	// check if user exists
	existedUser, err := c.userSvc.FindUserByID(r.Context(), targetUserID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if user is in the same merchant
	if existedUser.MerchantId != user.MerchantId {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, fmt.Errorf("user not found")))
		return
	}

	// check if role exists
	existedRole, errFindRole := c.roleSvc.FindRoleBySlug(r.Context(), payload.RoleSlug)
	if errFindRole != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, errFindRole))
		return
	}

	if user.UUID == targetUserID && strings.EqualFold(existedUser.Role.String, constant.RoleAdmin) && !strings.EqualFold(payload.RoleSlug, constant.RoleAdmin) {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrForbidden, constant.ErrForbiddenToChangeRole))
		return
	}

	isRoleChanged := payload.RoleSlug != existedUser.Role.String

	// fill user data
	userData := &userModel.User{
		UUID:             existedUser.UUID,
		Email:            strings.ToLower(payload.Email),
		Status:           existedUser.Status,
		Name:             payload.Name,
		Blocked:          existedUser.Blocked,
		MerchantId:       existedUser.MerchantId,
		RefreshToken:     existedUser.RefreshToken,
		IsChangePassword: existedUser.IsChangePassword,
		PinHash:          existedUser.PinHash,
		LastLoginAt:      existedUser.LastLoginAt,
		DeactivatedAt:    existedUser.DeactivatedAt,
		Password:         existedUser.Password,
		CreatedAt:        existedUser.CreatedAt,
		UpdatedAt:        time.Now().UTC(),
	}

	if isRoleChanged {
		userData.RefreshToken = sql.NullString{
			Valid: false,
		}
	}

	if err = c.userSvc.Update(r.Context(), userData); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if role is changed
	if isRoleChanged {
		// get userRole
		userRoleData, errFindUserRole := c.userRoleSvc.FindUserRoleByUserID(r.Context(), existedUser.UUID)
		if errFindUserRole != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, errFindUserRole))
			return
		}

		// assign role 'admin' to the user
		newUserRole := &userRole.UserRole{
			UUID:      userRoleData.UUID,
			UserID:    userRoleData.UserID,
			RoleID:    existedRole.UUID,
			CreatedAt: userRoleData.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		}

		if err = c.userRoleSvc.UpdateByUserID(r.Context(), newUserRole); err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}

		err = c.JWT.TerminateTokenOfUserRoleChanged(ctx, userData.Email)
		if err != nil {
			c.logger.Error(ctx, "failed to terminate the access token of user role changed", logger.Error(err), logger.String("userID", userData.UUID), logger.String("roleID", userRoleData.UUID))
		}

	}

	// assign role to response
	userData.Role = sql.NullString{String: existedRole.Slug, Valid: true}

	response.SendApiResponseOK(w, userData.ToResponse())
}
