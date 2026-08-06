package role

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Create		godoc
// @Summary		Create role endpoint
// @Description	Create role endpoint
// @ID			api-role-create
// @Tags		API - Role
// @Accept		json
// @Produce		json
// @Param		Request	body		role.CreateRoleRequest true "JSON Body for Create Role and Permissions"
// @Success		200  	{object}	response.ApiResponse{data=role.RoleMenuResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/roles/create [post]
// @Security	Bearer
func (c *RoleController) Create(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/role/Create")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload role.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	payload.MerchantID = user.MerchantId

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.roleSvc.CreateRoleAndPermissions(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, resp)
}
