package callback

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Get			godoc
// @Summary		To display a summary of the callback URLs
// @Description	To display a summary of the callback URLs
// @ID			settings-callbacks
// @Tags		Settings - Callbacks
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/settings/callbacks [get]
// @Security	Bearer
func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/callback/Get")
	defer segment.End()

	h.PerformCallbackSettingReqFunc(w, r, h.service.GetCallbackURLByMerchantId)
}

// GetApiKey	godoc
// @Summary		View callback API key (PIN required)
// @Description	View callback API key (PIN required)
// @ID			settings-callbacks-view-api-key
// @Tags		Settings - Callbacks
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=callback_model.CallbackAPIKeyResp}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/settings/callbacks/api-key [get]
// @Security	Bearer
func (h *handler) GetApiKey(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/callback/GetApiKey")
	defer segment.End()

	reqPIN := []byte{}
	if encoded := r.Header.Get(constant.HeaderXRequestPIN); encoded != "" {
		reqPIN, _ = base64.StdEncoding.DecodeString(encoded)
	}
	r = r.WithContext(context.WithValue(r.Context(), constant.CtxUserPINKey, string(reqPIN)))

	h.PerformCallbackSettingReqFunc(w, r, h.service.GetCallbackAPIKeyByMerchantId)
}

// Note: General function
func (h *handler) PerformCallbackSettingReqFunc(w http.ResponseWriter, r *http.Request, f func(context.Context, *callbackModel.CallbackURLSettingReq) (interface{}, error)) {
	ctx := r.Context()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &callbackModel.CallbackURLSettingReq{
		Info:       r,
		UserID:     user.UUID,
		MerchantID: user.MerchantId,
		MasterName: r.URL.Query().Get("masterName"),
	}
	if err := h.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := f(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
