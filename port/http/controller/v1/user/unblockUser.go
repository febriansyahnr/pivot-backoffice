package user

import (
	"encoding/json"
	"net/http"
	"strings"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UnblockUser		godoc
// @Summary			Unblock user because of entering wrong password and otp
// @Description		Unblock user because of entering wrong password and otp
// @ID				user-unblock
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param			Request	body		user.UserUnblockRequest true "JSON Body for Unblock User"
// @Success			200  	{object}	response.ApiResponse{data=string}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/users/unblock [post]
// @Security 		Bearer
func (c *UserController) UnblockUser(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/UnblockUser")
	defer segment.End()

	var (
		err error
	)

	var payload userModel.UserUnblockRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrValidation, err))
		return
	}

	for _, email := range payload.Email {
		// Check existed user by email
		errUnblock := c.userSvc.UnblockUser(ctx, strings.ToLower(email))
		if errUnblock != nil {
			response.SendApiResponseError(ctx, w, errUnblock)
			return
		}
	}

	response.SendApiResponseOK(w, "Unblock process success")
}
