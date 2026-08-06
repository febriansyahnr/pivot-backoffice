package payment

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// PaymentHistory		godoc
// @Summary		Payment history detail Endpoint
// @Description	Get Detail Payment Activity
// @ID			api-payment-history-detail
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.PaymentHistoryDetailResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/{id} [get]
// @Security	Bearer
func (c *PaymentController) PaymentHistory(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		paymentID = chi.URLParam(r, "payment_id")
	)
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payment/PaymentHistory")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if paymentID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	merchantID := user.MerchantId
	if r.URL.Query().Get("subMerchantId") != "" {
		err := c.merchantService.ValidateSubMerchantParent(ctx, user.MerchantId, r.URL.Query().Get("subMerchantId"))
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}

		merchantID = r.URL.Query().Get("subMerchantId")
	}

	result, err := c.paymentService.GetPaymentHistoryDetail(ctx, paymentModel.PaymentHistoryDetailOption{
		PaymentID:  paymentID,
		MerchantID: merchantID,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}

// GetEncryptionKey		godoc
// @Summary		Get Encryption Key Endpoint
// @Description	Get encryption key for payment data encryption
// @ID			api-payment-encryption-key
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.GetEncryptionKeyResponse}
// @Failure		401  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/encryption-key [get]
// @Security	Bearer
func (h *PaymentController) GetEncryptionKey(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetEncryptionKey")
	defer segment.End()

	if _, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims); !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	result, err := h.paymentService.GetEncryptionKey(ctx)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}
