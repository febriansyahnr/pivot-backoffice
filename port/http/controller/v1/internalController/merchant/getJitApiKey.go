package internal_merchant

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalMerchantController) GetJITApiKey(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchant/GetJITApiKey")
	defer segment.End()

	merchantId := chi.URLParam(r, "merchantId")
	_, err := uuid.Parse(merchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	jitKey, err := c.merchantSvc.GetOrGenerateJITApiKey(ctx, merchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, merchant.ToJITAPIKeyResponse(jitKey))
}
