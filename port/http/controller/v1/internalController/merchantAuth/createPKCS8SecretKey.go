package internalMerchantAuthController

import (
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// List				godoc
// @Summary			Create PKCS8 Secret Key
// @Description		List of all banks
// @ID				api-create-secret-key
// @Tags			API - Open Api
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.OpenApiResponse
// @Failure			500  	{object}	response.OpenApiResponse
// @Router			/internal/v1/secret-key [post]
// @Security		Bearer
func (c *InternalMerchantAuthController) CreatePKCS8SecretKey(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/CreatePKCS8SecretKey")
	defer segment.End()

	var (
		err error
	)

	merchantCtx, ok := r.Context().Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		err = fmt.Errorf("unauthorized")
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
	}

	if merchantCtx == nil {
		err = fmt.Errorf("unauthorized")
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}
	merchantID := merchantCtx.ID
	httputil.BindSubmerchantID(r, &merchantID)

	result, err := c.merchantSvc.CreatePKCS8SecretKey(r.Context(), merchantID)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, result)
}
