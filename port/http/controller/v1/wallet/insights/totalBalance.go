package walletInsightsController

import (
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *WalletInsightsController) TotalBalance(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		isRefresh = false
	)
	_, segment := otelTracer.Start(ctx, "port/http/controller/v1/wallet/insights/TotalBalance")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	refresh := r.URL.Query().Get("refresh")
	if strings.ToLower(refresh) != "true" {
		isRefresh = true
	}

	data, err := c.insightSvc.TotalBalance(r.Context(), user.MerchantId, isRefresh)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, data)
}
