package crmProductController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMProductController) AddMerchantSelectedProduct(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		payload product.AddMerchantProductRequest
	)
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/product/AddMerchantSelectedProduct")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	err := c.productService.AddMerchantSelectedProduct(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, nil)
}
