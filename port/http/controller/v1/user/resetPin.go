package user

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ResetPIN 	godoc
// @Summary		ResetPIN endpoint is used to create a new PIN through the forgot PIN process.
// @Description	ResetPIN endpoint is used to create a new PIN through the forgot PIN process.
// @ID			api-users-reset-pin
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param 		Request	body			user.ResetPinRequest true "JSON Body for Reset PIN"
// @Success		200  	{object}		response.ApiResponse{data=user.UpdateStatusResp}
// @Failure		500  	{object}		response.ApiResponse
// @Router		/api/v1/auth/reset-pin	[patch]
// @Security 	Bearer
func (c *UserController) ResetPIN(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/ResetPIN")
	defer segment.End()

	claims, _ := ctx.Value(constant.CtxTokenOTPKey).(*otpModel.TokenOTPClaims)

	var payload userModel.ResetPinRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.userSvc.ResetPIN(ctx, claims.UUID, payload.Pin); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, userModel.UpdateStatusResp{Updated: true})
}
