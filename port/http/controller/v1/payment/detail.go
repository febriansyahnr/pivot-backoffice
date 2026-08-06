package payment

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPaymentDetailForPaymentUI		godoc
// @Summary		Payment detail for payment UI endpoint
// @Description	Payment detail for payment UI endpoint
// @ID			api-payment-detail-for-payment-ui
// @Tags		API - Payment UI
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.PaymentDetailForPaymentUIResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/detail [get]
// @Security	Bearer
func (c *PaymentController) GetPaymentDetailForPaymentUI(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetPaymentDetailForPaymentUI")
	defer segment.End()

	// Get PaymentID from token
	paymentID, ok := ctx.Value(constant.CtxPaymentID).(string)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	// Get payment detail
	paymentDetail, err := c.paymentService.GetPaymentDetailForPaymentUI(ctx, paymentID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, paymentDetail)
}

func (c *PaymentController) GetPaymentImages(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetPaymentImages")
	defer segment.End()

	_, ok := ctx.Value(constant.CtxPaymentID).(string)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	images, err := c.paymentService.GetImages(ctx)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, images)
}

func (c *PaymentController) GetPaymentInstructions(w http.ResponseWriter, r *http.Request) {

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetPaymentInstructions")
	defer segment.End()

	_, ok := ctx.Value(constant.CtxPaymentID).(string)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	paymentMethod := r.URL.Query().Get("paymentMethod")

	instructions, err := c.paymentService.GetPaymentInstructions(ctx, paymentMethod)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, instructions)
}
