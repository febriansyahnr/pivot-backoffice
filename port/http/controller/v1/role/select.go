package role

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList		godoc
// @Summary		List roles endpoint
// @Description	List roles endpoint
// @ID			api-role-list
// @Tags		API - Role
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=[]role.RoleResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/roles/list [get]
// @Security	Bearer
func (c *RoleController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/GetList")
	defer segment.End()

	var (
		page    int64 = 1
		perPage int64 = constant.DefaultPaginationPageSize
		err     error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		err = constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	// Get query params
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")

	// Validation and parsing
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	if perPageStr != "" {
		perPage, err = strconv.ParseInt(perPageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}

	filter := &roleModel.GetRoleFilterRequest{
		MerchantID: user.MerchantId,
	}

	list, err := c.roleSvc.GetList(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
