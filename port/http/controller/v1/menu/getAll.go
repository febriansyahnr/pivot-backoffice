package menuController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetAll				godoc
// @Summary				Get all menus
// @Description			Get all menus
// @ID					api-menus
// @Tags				API - Menu
// @Accept				json
// @Produce				json
// @Param				forRoleManagement	query		bool	false	"When true, exclude Home menu (used in role management UI)"
// @Success				200  		{object}	response.ApiResponse
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/menus [get]
// @Security			Bearer
func (c *MenuController) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/menu/GetAll")
	defer segment.End()

	// Check if this is for role management UI (exclude auto-granted menus like Home)
	forRoleManagement := r.URL.Query().Get("forRoleManagement") == "true"

	// Get all menus
	menus, err := c.menuSvc.GetAll(ctx, forRoleManagement)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, menus)
}
