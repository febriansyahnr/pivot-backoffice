package crmProductController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMProductController) UpdateMerchantProductAvailability(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		payload product.UpdateMerchantSelectedProductAvailabilityRequest
	)
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/product/UpdateMerchantProductAvailability")
	defer segment.End()

	merchantId := chi.URLParam(r, "merchantId")
	_, err := uuid.Parse(merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	payload.MerchantID = merchantId
	if err := c.validate.Struct(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	err = c.productService.UpdateMerchantProductAvailability(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, nil)
}
