package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMUserController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/crmController/v1/user/Update")
	defer segment.End()

	var (
		err error
	)

	id := chi.URLParam(r, "user_id")
	if id == "" {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, fmt.Errorf("user id is required")))
		return
	}

	var payload userModel.CRMUserUpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	// check if user exists
	existedUser, err := c.userSvc.FindUserByID(r.Context(), id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// fill user data
	userData := &userModel.User{
		UUID:             existedUser.UUID,
		Email:            strings.ToLower(payload.Email),
		Status:           payload.Status,
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
	if err = c.userSvc.Update(r.Context(), userData); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if role exists
	existedRole, errFindRole := c.roleSvc.FindRoleBySlug(r.Context(), payload.RoleSlug)
	if errFindRole != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrInternal, errFindRole))
		return
	}

	// check if role is changed
	if payload.RoleSlug != existedUser.Role.String {
		// get userRole
		userRoleData, errFindUserRole := c.userRoleSvc.FindUserRoleByUserID(r.Context(), existedUser.UUID)
		if errFindUserRole != nil {
			response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrInternal, errFindUserRole))
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
	}

	// assign role to response
	userData.Role = sql.NullString{String: existedRole.Slug, Valid: true}

	// publish activity when user status is updated
	if payload.Status == constant.UserStatusBlocked {
		// Publish activity, do nothing on error
		_ = c.rabbitMqExt.PublishActivity(
			ctx,
			&userData.MerchantId,
			&userData.UUID,
			constant.TagAccount,
			constant.ActivityUserBlockedByOps,
			map[string]string{
				"email":    userData.Email,
				"password": "********",
			},
		)
	}

	response.SendApiResponseOK(w, userData.ToResponse())
}
