package role

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

// Update		godoc
// @Summary		Update role endpoint
// @Description	Update role endpoint
// @ID			role-update
// @Tags		API - Role
// @Accept		json
// @Produce		json
// @Param 		role_id	path		string true "Role ID to update role"
// @Param		Request	body		role.UpdateRoleRequest true "JSON Body for Update Role"
// @Success		200  	{object}	response.ApiResponse{data=role.RoleMenuResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/roles/{role_id} [put]
// @Security 	Bearer
func (c *RoleController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/role/Update")
	defer segment.End()

	var (
		err error
	)

	id := chi.URLParam(r, "role_id")
	if err = uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("role id is required")))
		return
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload roleModel.UpdateRoleRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}
	payload.MerchantID = user.MerchantId
	payload.RoleID = id

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Update role, menu & permissions
	resp, err := c.roleSvc.UpdateRoleAndPermissions(r.Context(), &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
