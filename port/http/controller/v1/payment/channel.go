package payment

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetChannelList		godoc
// @Summary		Payment Channel List Endpoint
// @Description	Gather list of payment channels available for payment
// @ID			api-payment-channel-list
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=[]*paymentModel.PaymentMethodWithPivot}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/channels [get]
// @Security	Bearer
func (c *PaymentController) GetChannelList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetChannelList")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID := user.MerchantId

	// when derived merchant ID is set, use it as the merchant ID
	// non-kyc merchant will use the parent merchant ID as source of truth
	// this feature should changed when the merchant able to change sub-merchant payment method
	derivedMerchantID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string)
	if ok && derivedMerchantID != "" {
		merchantID = derivedMerchantID
	}

	methods, err := c.paymentMethodService.GetPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
		Category:   constant.TypePayment,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, methods)
}

// GetChannelDocuments		godoc
// @Summary		Payment Channel Documents Endpoint
// @Description	collection information about documents required for payment channels
// @ID			api-payment-channel-documents
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.PaymentDetailForPaymentUIResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/channels [get]
// @Security	Bearer
func (c *PaymentController) GetChannelDocuments(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetChannelDocuments")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	documents, err := c.paymentMethodService.GetRequiredMerchantDocuments(ctx, &paymentMethodModel.GetRequiredMerchantDocumentsRequest{
		MerchantID:      user.MerchantId,
		PaymentMethodID: r.PathValue("paymentMethodId"),
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, documents)
}

func (c *PaymentController) UpdatePaymentMethodStatus(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/UpdatePaymentMethodStatus")
	defer segment.End()

	var (
		payload paymentModel.UpdatePaymentMethodStatusRequest
	)

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	err = c.validate.Struct(payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	request := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID: r.PathValue("paymentMethodId"),
		},
		MerchantID: user.MerchantId,
		IsActive:   *payload.Status,
	}

	if *payload.Status {
		err = c.paymentMethodService.Activate(ctx, request)
	} else {
		err = c.paymentMethodService.Deactivate(ctx, request)
	}

	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]interface{}{
		"merchantId":      user.MerchantId,
		"paymentMethodId": request.PaymentMethod.UUID,
		"status":          payload.Status,
	})
}

// GetChannelListWithPaymentToken		godoc
// @Summary		Payment Channel List Endpoint (with Payment Token)
// @Description	Gather list of payment channels available for payment using payment token
// @ID			api-payment-channel-list-token
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=[]*paymentModel.PaymentMethodWithPivot}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/payments/channels [get]
// @Param		token	query	string	true	"Payment token"
func (c *PaymentController) GetChannelListWithPaymentToken(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetChannelListWithPaymentToken")
	defer segment.End()

	// Get payment ID from context (set by PaymentTokenMiddleware)
	paymentID, ok := ctx.Value(constant.CtxPaymentID).(string)
	if !ok || paymentID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	// Get payment details to extract merchant ID
	paymentDetail, err := c.paymentService.GetPaymentDetailForPaymentUI(ctx, paymentID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	merchant, err := c.merchantService.FindMerchantByID(ctx, paymentDetail.MerchantID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	merchantID := merchant.UUID
	if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
		merchantID = merchant.ParentID.String
	}

	// Get payment methods for this merchant
	methods, err := c.paymentMethodService.GetPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
		Category:   constant.TypePayment,
		Payment:    paymentDetail,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, methods)
}
