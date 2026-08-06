package user

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CreatePIN	godoc
// @Summary		Create PIN endpoint
// @Description	Create PIN endpoint
// @ID			user-create-pin
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.CreatePinRequest true "JSON Body for Create User"
// @Success		200  	{object}	response.ApiResponse{data=user.CreatePinRequest}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/pin [post]
// @Security 	Bearer
func (c *UserController) CreatePin(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/CreatePin")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	var payload userModel.CreatePinRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.userSvc.CreatePin(ctx, user.UUID, payload.Pin); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, payload)
}
