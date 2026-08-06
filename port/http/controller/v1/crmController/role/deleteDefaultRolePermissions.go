package role

import (
	"encoding/json"
	"net/http"

	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// DeleteDefaultRolePermissions godoc
// @Summary		CRM - Delete permissions from default roles
// @Description	Delete specific permissions from default roles (ADMIN, MAKER, APPROVER, DEVELOPER, OPERATION, CROSSBORDER_OPERATOR, PLATFORM). Only removes the specified permissions, all others remain unchanged.
// @ID			crm-role-delete-default-permissions
// @Tags		CRM - Role
// @Accept		json
// @Produce		json
// @Param		Request	body		role.CRMUpdateDefaultRolePermissionsRequest true "JSON Body specifying which permissions to delete"
// @Success		200  	{object}	response.ApiResponse{data=role.RoleMenuResponse}
// @Failure		400  	{object}	response.ApiResponse
// @Failure		422  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/crm/roles/default-permissions [delete]
// @Security 	Bearer
func (c *CRMRoleController) DeleteDefaultRolePermissions(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/role/DeleteDefaultRolePermissions")
	defer segment.End()

	var payload roleModel.CRMUpdateDefaultRolePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Delete default role permissions
	resp, err := c.roleSvc.DeleteDefaultRolePermissions(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
