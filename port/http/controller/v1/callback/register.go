package callbackController

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RegisterCallback		godoc
// @Summary				Register callback URL
// @Description			Register callback URL
// @ID					api-callbacks-register-callback
// @Tags				API - Callbacks
// @Accept				json
// @Produce				json
// @Param				Request	body		callback_model.RegisterCallbackRequest true "JSON Body for Register Callback"
// @Success				200  	{object}	response.ApiResponse
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/callbacks/ [post]
// @Security			Bearer
func (c *CallbackController) RegisterCallback(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/callback/RegisterCallback")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Validate Payload
	var payload callbackModel.RegisterCallbackRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Convert merchantID to uuid
	payload.MerchantID, err = uuid.Parse(user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	// Hit Register Callback
	callback, err := c.callbackSvc.RegisterCallback(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagCallback,
		constant.ActivityUserRegisterCallback,
		payload,
	)

	response.SendApiResponseOK(w, callback)
}
