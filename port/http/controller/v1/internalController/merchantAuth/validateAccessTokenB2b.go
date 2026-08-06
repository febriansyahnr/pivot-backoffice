package internalMerchantAuthController

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalMerchantAuthController) ValidateAccessTokenB2b(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/ValidateAccessTokenB2b")
	defer segment.End()

	var (
		err     error
		payload merchantModel.ValidateAccessTokenB2bRequest
	)

	merchantId := r.Header.Get(constant.HeaderXMerchantId)
	payload.MerchantId = merchantId

	// Get token from authorization header
	tokenHeader := r.Header.Get(constant.HeaderAuthorization)
	if !strings.Contains(tokenHeader, "Bearer") {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("invalid token")))
		return
	}
	tokenString := strings.Replace(tokenHeader, "Bearer ", "", -1)
	payload.AccessToken = tokenString

	claims, err := c.merchantSvc.ValidateAccessTokenB2b(r.Context(), &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, claims)
}

func (c *InternalMerchantAuthController) ValidateSNAPAccessTokenB2b(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/ValidateAccessTokenB2b")
	defer segment.End()

	var (
		err     error
		payload merchantModel.ValidateAccessTokenB2bRequest
	)

	// Get token from authorization header
	tokenHeader := r.Header.Get(constant.HeaderAuthorization)
	if !strings.Contains(tokenHeader, "Bearer") {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("invalid token")))
		return
	}
	tokenString := strings.Replace(tokenHeader, "Bearer ", "", -1)
	payload.AccessToken = tokenString
	payload.IsSnapRequest = true

	claims, err := c.merchantSvc.ValidateAccessTokenB2b(r.Context(), &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, claims)
}
