package callback

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// TestAndSaveCallbackURL	godoc
// @Summary					Endpoint to define, edit, and test callback URLs
// @Description				Endpoint to define, edit, and test callback URLs
// @ID						settings-manage-callback-urls
// @Tags					Settings - Callbacks
// @Accept					json
// @Produce					json
// @Param					Request	body		callback_model.TestAndSaveCallbackURLReq true "JSON Body for test and save callback url"
// @Success					200  	{object}	response.ApiResponse{data=callback_model.TestAndSaveCallbackURLResp}
// @Failure					500  	{object}	response.ApiResponse
// @Router					/api/v1/settings/callbacks/urls/:master_id [post]
// @Security				Bearer
func (h *handler) TestAndSaveCallbackURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/setting/callback/TestAndSaveCallbackURL")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := callbackModel.TestAndSaveCallbackURLReq{
		CallbackMasterID: r.PathValue("master_id"),
		MerchantID:       user.MerchantId,
		UserID:           user.UUID,
		Info:             r,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.TestAndSaveCallbackURL(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}

func (h *handler) TestAndSaveSnapB2b(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/setting/callback/TestAndSaveSnapB2b")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := callbackModel.TestAndSaveCallbackURLReq{
		CallbackMasterID: r.PathValue("master_id"),
		MerchantID:       user.MerchantId,
		UserID:           user.UUID,
		Info:             r,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.TestAndSaveB2b(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}

func (h *handler) TestAndSaveSnapPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/setting/callback/TestAndSaveSnapPayment")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := callbackModel.TestAndSaveCallbackURLReq{
		CallbackMasterID: r.PathValue("master_id"),
		MerchantID:       user.MerchantId,
		UserID:           user.UUID,
		Info:             r,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.TestAndSaveSnapPayment(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
