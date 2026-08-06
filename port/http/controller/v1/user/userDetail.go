package user

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UserDetail		godoc
// @Summary			User Detail endpoint
// @Description		User Detail endpoint
// @ID				api-user-detail
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param 			email	path		string true "User Email"
// @Success			200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/user-detail [get]
// @Security 		Bearer
func (c *UserController) UserDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/UserDetail")
	defer segment.End()

	var (
		err error
	)

	// Get user info from jwt
	userInfo := r.Context().Value(constant.CtxUserInfoKey)
	user := userInfo.(*userModel.UserTokenClaims)

	userDetail, err := c.userSvc.UserDetail(r.Context(), user.UUID)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))

		return
	}

	response.SendApiResponseOK(w, userDetail)
}
