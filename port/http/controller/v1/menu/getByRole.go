package menuController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetByRole			godoc
// @Summary				Get menus by role
// @Description			Get menus by role
// @ID					api-menus-by-role
// @Tags				API - Menu
// @Accept				json
// @Produce				json
// @Success				200  		{object}	response.ApiResponse
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/menus/role [get]
// @Security			Bearer
func (c *MenuController) GetByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/menu/GetByRole")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	userRole, err := c.userRoleSvc.FindUserRoleByUserID(ctx, user.UUID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Get menu by role
	menus, err := c.menuSvc.GetByRole(ctx, userRole.RoleID, true)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, menus)
}
