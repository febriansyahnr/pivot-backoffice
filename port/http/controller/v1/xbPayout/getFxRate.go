package xbPayoutController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *xbPayoutController) GetFxRate(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetFxRate")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	sourceCurrency := r.URL.Query().Get("sourceCurrency")
	if sourceCurrency == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidSourceCurrency))
		return
	}

	destinationCurrency := r.URL.Query().Get("destinationCurrency")
	if destinationCurrency == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidDestinationCurrency))
		return
	}

	resp, err := c.xbPayoutSvc.GetFxRate(ctx, &xbModel.GetFxRateRequest{
		MerchantId:          user.MerchantId,
		SourceCurrency:      sourceCurrency,
		DestinationCurrency: destinationCurrency,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, resp)
}
