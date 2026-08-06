package internal_merchant

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalMerchantController) Detail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchant/Detail")
	defer segment.End()

	merchantId := chi.URLParam(r, "id")
	_, err := uuid.Parse(merchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}
	if merchant == nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
		return
	}

	response.SendApiResponseOK(w, merchant.ToResponse())
}
