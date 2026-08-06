package user

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UserProfile		godoc
// @Summary			User Profile endpoint
// @Description		Get current authenticated user profile
// @ID				api-user-profile
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/user/profile [get]
// @Security 		Bearer
func (c *UserController) UserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/UserProfile")
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
