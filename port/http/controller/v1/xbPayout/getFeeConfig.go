package xbPayoutController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *xbPayoutController) GetFeeConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetFeeConfig")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	ctx, configs, err := c.merchantSvc.GetMerchantIdForConfigs(ctx, user.MerchantId, false)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	resp, err := c.xbPayoutSvc.GetFeeConfig(ctx, configs.MerchantTransactionConfig)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
