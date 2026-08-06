package v2InternalUnifiedPaymentController

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) Capture(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/Capture")
	defer segment.End()

	var (
		err  error
		resp *unifiedPaymentModel.CaptureResponse
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

	paymentID := chi.URLParam(r, "uuid")
	if err = uuid.Validate(paymentID); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	payload := &unifiedPaymentModel.CaptureRequest{
		PaymentID:  paymentID,
		MerchantID: merchantID,
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate capture payload
	if errValidate := c.validateCapturePayload(payload); errValidate != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errValidate)
		return
	}

	// Call service to capture payment
	if resp, err = c.unifiedPaymentSvc.Capture(ctx, payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *paymentController) validateCapturePayload(payload *unifiedPaymentModel.CaptureRequest) error {
	// If releaseRemainingAmount is false, amount is mandatory
	if !payload.ReleaseRemainingAmount && payload.Amount == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("amount is required when releaseRemainingAmount is false"))
	}

	// If releaseRemainingAmount is true, amount is optional
	// If amount is provided, it must be greater than 0
	if payload.Amount != nil && payload.Amount.Value <= 0 {
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("amount value must be greater than 0"))
	}

	return nil
}
