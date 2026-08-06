package menuController

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetByRoleId			godoc
// @Summary				Get menus by role_id
// @Description			Get menus by role_id
// @ID					api-menus-by-role-id
// @Tags				API - Menu
// @Accept				json
// @Produce				json
// @Param 				roleID		path		string true "Role ID"
// @Success				200  		{object}	response.ApiResponse
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/menus/role/:id [get]
// @Security			Bearer
func (c *MenuController) GetByRoleId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/menu/GetByRoleId")
	defer segment.End()

	id := chi.URLParam(r, "role_id")
	if errValidate := uuid.Validate(id); errValidate != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("role id is required")))
		return
	}

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// check if role is existed
	role, err := c.roleSvc.FindRoleById(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if user merchant is same with role merchant
	if role.Type == constant.RoleTypeCustom && user.MerchantId != role.MerchantID.String {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserUnauthorized))
		return
	}

	// Get menu by role
	menus, err := c.menuSvc.GetByRole(ctx, role.UUID, false)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, menus)
}
