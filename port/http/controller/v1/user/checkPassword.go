package user

import (
	"encoding/json"
	e "errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CheckCurrentPassword	godoc
// @Summary				Check Password endpoint
// @Description			Check Password endpoint
// @ID					api-user-check-password
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param				Request	body		user.CheckPasswordRequest true "JSON Body for Check Password"
// @Success				200  	{object}	response.ApiResponse{data=nil}
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/users/check-password [patch]
// @Security			Bearer
func (c *UserController) CheckCurrentPassword(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/CheckCurrentPassword")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.CheckPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrValidation, err))
		return
	}

	if err := c.userSvc.CheckCurrentPassword(r.Context(), user.UUID, payload.Password); err != nil {
		if e.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
			// publish activity, do nothing on error
			_ = c.rabbitMqExt.PublishActivity(
				ctx,
				&user.MerchantId,
				&user.UUID,
				constant.TagAccount,
				constant.ErrFailedCheckPassword.Error(),
				map[string]string{
					"email":    user.Email,
					"password": "********",
				},
			)
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
