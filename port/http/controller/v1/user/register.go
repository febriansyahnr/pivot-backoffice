package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Register		godoc
// @Summary		Register endpoint with email and password
// @Description	Register endpoint with email and password
// @ID			api-user-register
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserRegisterRequest true "JSON Body for Register"
// @Success		200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/auth/register [post]
func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/Register")
	defer segment.End()

	var (
		err error
	)

	var payload userModel.UserRegisterRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	user := &userModel.User{
		UUID:             uuid.New().String(),
		Email:            strings.ToLower(payload.Email),
		Status:           constant.UserStatusActive,
		Name:             payload.Name,
		IsChangePassword: 1,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// hash password
	user.Password = user.EncryptPassword(payload.Password)

	if err = c.userSvc.Create(r.Context(), user); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if role 'admin' exists
	existedRole, err := c.roleSvc.FindRoleBySlug(r.Context(), constant.RoleAdmin)
	if err != nil && strings.Contains(err.Error(), "not found") {
		newRole := &role.Role{
			UUID:      uuid.New().String(),
			Name:      "Admin",
			Slug:      "admin",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err = c.roleSvc.Create(r.Context(), newRole); err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}

		existedRole = newRole
	} else if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// assign role 'admin' to the user
	newUserRole := &userRole.UserRole{
		UUID:      uuid.New().String(),
		UserID:    user.UUID,
		RoleID:    existedRole.UUID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err = c.userRoleSvc.Create(r.Context(), newUserRole); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// generate refresh token
	refreshToken, err := c.JWT.GenerateRefreshToken(r.Context(), user, time.Now().UTC().Add(constant.REFRESH_EXPIRATION_DURATION))
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// update user
	user.Role = sql.NullString{String: existedRole.Slug, Valid: true}
	user.RefreshToken = sql.NullString{String: refreshToken, Valid: true}
	if err = c.userSvc.Update(r.Context(), user); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, user.ToResponse())
}
