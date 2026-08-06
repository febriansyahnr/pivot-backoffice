package user

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ListUsers	godoc
// @Summary		List users endpoint
// @Description	List users endpoint
// @ID			internal-user-list
// @Tags		Internal - User
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=[]user.UserResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/internal/v1/users/list [get]
func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/ListUsers")
	defer segment.End()

	var (
		err   error
		users []*userModel.User
	)

	if users, err = c.userSvc.ListUsers(ctx, 10, 0); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, toListUsersResponse(users))
}

func toListUsersResponse(users []*userModel.User) []*userModel.UserResponse {
	var res []*userModel.UserResponse
	for _, user := range users {
		res = append(res, user.ToResponse())
	}
	return res
}

// FindByID	godoc
// @Summary		Get users by ID endpoint
// @Description	Get users by ID endpoint
// @ID			api-get-user-by-id
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param 		user_id	path		string true "User ID to get user"
// @Success		200  	{object}	response.ApiResponse{data=[]user.UserResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/{user_id} [get]
func (c *UserController) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/FindByID")
	defer segment.End()

	id := chi.URLParam(r, "user_id")
	if id == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("user id is required")))
		return
	}

	user, err := c.userSvc.FindUserByID(r.Context(), id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, user.ToResponse())
}
