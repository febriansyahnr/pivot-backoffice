package v1InternalUnifiedPaymentController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/Update")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	payload := paymentModel.UpdateUnifiedPaymentRequest{
		MerchantID: merchantAuth.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err := c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if c.isPaymentMigrationV1ToV2Enabled(ctx, payload.MerchantID) {
		response.SendOpenApiResponseOK(w, map[string]any{
			"message":     "This API endpoint has been deprecated and is no longer available.",
			"deprecated":  true,
			"alternative": "/v2/payments",
		})
		return
	}

	resp, err := c.paymentSvc.UpdateUnifiedPayment(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
