package internalXbController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalXbController) GetFxRate(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetFxRate")
	defer segment.End()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	sourceCurrency := r.URL.Query().Get("sourceCurrency")
	if sourceCurrency == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidSourceCurrency))
		return
	}

	destinationCurrency := r.URL.Query().Get("destinationCurrency")
	if destinationCurrency == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidDestinationCurrency))
		return
	}

	resp, err := c.xbPayoutSvc.GetFxRate(ctx, &xbModel.GetFxRateRequest{
		MerchantId:          merchantID,
		SourceCurrency:      sourceCurrency,
		DestinationCurrency: destinationCurrency,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}
