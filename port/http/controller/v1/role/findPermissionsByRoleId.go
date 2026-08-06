package role

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FindPermissionsByRoleId		godoc
// @Summary						Find permission by Role ID
// @Description					Find permission by Role ID
// @ID							api-permission-find-by-role-id
// @Tags						API - Permission
// @Accept						json
// @Produce						json
// @Param 						roleID	path		string true "Role ID"
// @Success						200  		{object}	response.ApiResponse{data=merchant.MerchantResponse}
// @Failure						500  		{object}	response.ApiResponse
// @Router						/api/v1/roles/:id/permissions [get]
// @Security					Bearer
func (c *RoleController) FindPermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/role/FindPermissionsByRoleId")
	defer segment.End()

	var (
		err         error
		permissions []*permissionModel.Permission
	)

	id := chi.URLParam(r, "role_id")
	if errValidate := uuid.Validate(id); errValidate != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("role id is required")))
		return
	}

	if permissions, err = c.permissionSvc.FindByRoleId(r.Context(), id); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, permissions)
}
