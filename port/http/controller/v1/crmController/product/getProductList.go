package crmProductController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMProductController) GetProductList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/product/paymentMethod/GetByMerchant")
	defer segment.End()

	productList, err := c.productService.GetProductList(ctx)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, productList)
}
