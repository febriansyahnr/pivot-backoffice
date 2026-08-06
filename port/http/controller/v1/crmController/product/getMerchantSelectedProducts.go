package crmProductController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMProductController) GetMerchantSelectedProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/product/GetMerchantSelectedProducts")
	defer segment.End()

	merchantId := chi.URLParam(r, "merchantId")
	_, err := uuid.Parse(merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID))
		return
	}

	productList, err := c.productService.GetMerchantSelectedProducts(ctx, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, productList)
}
