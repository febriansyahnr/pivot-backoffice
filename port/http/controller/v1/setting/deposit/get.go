package depositSettingController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Get			godoc
// @Summary		Get deposit settings
// @Description	Get a list of configurations and values ​​set in deposit settings
// @ID			settings-deposit-view
// @Tags		Settings - Deposit
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=merchant.DepositSettingResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/settings/deposit [get]
// @Security	Bearer
func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/deposit/Get")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		user.MerchantId = subMerchantId
	}

	if resp, err := h.merchantSvc.GetDepositSetting(ctx, user.MerchantId); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
