package v2InternalUnifiedPaymentController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) SimulatePayment(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/SimulatePayment")
	defer segment.End()

	var (
		err error
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)
	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	request := &unifiedPaymentModel.SimulatePaymentRequest{
		MerchantID: merchantID,
	}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err = c.validate.Struct(request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	if err = c.unifiedPaymentSvc.SimulatePayment(ctx, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, map[string]any{
		"updated": true,
	})
}
