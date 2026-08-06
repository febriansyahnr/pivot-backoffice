package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// AddUserRole	godoc
// @Summary		Add user role endpoint
// @Description	Add user role endpoint
// @ID			internal-role-add-user-role
// @Tags		Internal - Role
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserAddRoleRequest true "JSON Body for Assign Role"
// @Success		200  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/roles/assign [post]
func (c *UserController) AddUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/AddUserRole")
	defer segment.End()

	var (
		err error
	)

	var payload userModel.UserAddRoleRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// get user detail by email
	user, err := c.userSvc.FindUserByEmail(r.Context(), payload.Email)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if user == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, fmt.Errorf("user not found")))
		return
	}

	// get role by slug
	role, err := c.roleSvc.FindRoleBySlug(r.Context(), payload.RoleSlug)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if role == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, fmt.Errorf("role not found")))
		return
	}

	// add role to user
	req := userRole.UserRole{
		UUID:      uuid.NewString(),
		UserID:    user.UUID,
		RoleID:    role.UUID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	errAssign := c.userRoleSvc.Create(r.Context(), &req)
	if errAssign != nil {
		response.SendApiResponseError(ctx, w, errAssign)
		return
	}

	response.SendApiResponseOK(w, "Ok")
}
