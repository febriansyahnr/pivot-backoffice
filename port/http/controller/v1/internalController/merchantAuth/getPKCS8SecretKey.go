package internalMerchantAuthController

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// List				godoc
// @Summary			Create PKCS8 Secret Key
// @Description		List of all banks
// @ID				api-get-secret-key
// @Tags			API - Open Api
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.Response
// @Failure			500  	{object}	response.Response
// @Router			/internal/v1/secret-key [get]
// @Security		Bearer
func (c *InternalMerchantAuthController) GetPKCS8SecretKey(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/GetPKCS8SecretKey")
	defer segment.End()

	var (
		err error
	)

	merchantID := r.Header.Get(constant.ClientIdKey)
	if merchantID == "" {
		err = fmt.Errorf("unauthorized")
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	// get PKCS8 secret key
	secretKey, err := c.merchantSvc.GetPKCS8SecretKey(ctx, merchantID)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, secretKey)
}

// encrypting utils function, will deprecate soon :)
func (c *InternalMerchantAuthController) UtilEncryptingKey(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/UtilEncryptingKey")
	defer segment.End()

	type payloadUtil struct {
		Key  string `json:"key"`
		Data string `json:"data"`
	}

	var (
		err     error
		payload payloadUtil
	)

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	encrypted, err := c.merchantSvc.UtilEncryptingKey(ctx, payload.Key, payload.Data)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, encrypted)
}
