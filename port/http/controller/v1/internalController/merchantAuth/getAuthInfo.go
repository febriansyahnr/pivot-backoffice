package internalMerchantAuthController

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalMerchantAuthController) GetAuthInfo(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/GetAuthInfo")
	defer segment.End()

	const merchantNotFound = "merchant not found"
	var (
		err error
	)

	// Merchant info from JWT
	merchantInfoFromCtx := ctx.Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		err = errors.New(merchantNotFound)
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	merchantInfo, err := c.merchantSvc.FindMerchantByID(ctx, merchantCtx.MerchantId)
	if err != nil {
		err = errors.New(merchantNotFound)
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	if merchantInfo == nil {
		err = errors.New(merchantNotFound)
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	response.SendOpenApiResponseOK(w, merchantInfo)
}
