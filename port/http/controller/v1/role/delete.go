package role

import (
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Delete		godoc
// @Summary		Delete role endpoint
// @Description	Delete role endpoint
// @ID			api-role-delete
// @Tags		API - Role
// @Accept		json
// @Produce		json
// @Param 		role_id	path		string true "Role ID to delete role"
// @Success		200  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/roles/{role_id} [delete]
// @Security 	Bearer
func (c *RoleController) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/role/Delete")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	roleID := chi.URLParam(r, "role_id")
	if err := uuid.Validate(roleID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("role id is required")))
		return
	}

	if err := c.roleSvc.Delete(ctx, user.MerchantId, roleID); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, map[string]interface{}{"deleted": true})
}
