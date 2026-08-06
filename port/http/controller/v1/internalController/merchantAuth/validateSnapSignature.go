package internalMerchantAuthController

import (
	"encoding/json"
	"net/http"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalMerchantAuthController) ValidateSNAPSignature(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		payload merchantModel.ValidateSnapSignatureRequest
	)

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/ValidateSNAPSignature")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(responseHttp.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(responseHttp.HttpErrRequest, err))
		return
	}

	err = c.merchantSvc.ValidateSnapRequestSignature(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, "ok")
}

func (c *InternalMerchantAuthController) GenerateSNAPSignature(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		payload merchantModel.GenerateSnapSignatureRequest
	)

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/GenerateSNAPSignature")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(responseHttp.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(responseHttp.HttpErrRequest, err))
		return
	}

	signature, err := c.merchantSvc.GenerateSnapRequestSignature(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}
	response.SendOpenApiResponseOK(w, signature)
}
