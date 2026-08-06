package depositSettingController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// SetAutoWithdrawalStatus	godoc
// @Summary					Auto withdrawal settings
// @Description				Auto withdrawal settings
// @ID						settings-deposit-auto-withdrawal
// @Tags					Settings - Deposit
// @Accept					json
// @Produce					json
// @Success					200  	{object}	response.ApiResponse{data=merchant.AutoWithdrawalSettingRequest}
// @Failure					500  	{object}	response.ApiResponse
// @Router					/api/v1/settings/deposit/auto-withdrawal [patch]
// @Security				Bearer
func (h *handler) SetAutoWithdrawal(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/deposit/SetAutoWithdrawal")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &merchant.AutoWithdrawalSettingRequest{
		UserId:     user.UUID,
		MerchantId: user.MerchantId,
	}
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		request.MerchantId = subMerchantId
	}

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.merchantSvc.SetAutoWithdrawal(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, request)
}
