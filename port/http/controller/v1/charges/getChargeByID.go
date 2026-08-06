package charges

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetChargeByID		godoc
// @Summary		Get charge by ID Endpoint
// @Description	Get charge details by ID
// @ID			api-charges-get-by-id
// @Tags		API - Charges
// @Accept		json
// @Produce		json
// @Param		uuid	path		string	true	"Charge UUID"
// @Success		200  	{object}	response.ApiResponse{data=unifiedPaymentModel.ChargeResponse}
// @Failure		400  	{object}	response.ApiResponse
// @Failure		401  	{object}	response.ApiResponse
// @Failure		404  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/charges/{uuid} [get]
// @Security	Bearer
func (c *ChargesController) GetChargeByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/charges/GetChargeByID")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	chargeID := chi.URLParam(r, "uuid")
	if chargeID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err := uuid.Validate(chargeID); err != nil {
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

		// Set parent merchant ID in context for sub-merchant requests
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, user.MerchantId)
		merchantID = r.URL.Query().Get("subMerchantId")
	}

	request := unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
		ChargeID:   chargeID,
		MerchantID: merchantID,
	}

	result, err := c.unifiedPaymentService.GetChargeDetail(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if result.ChargePaymentMethodDetails != nil && result.VirtualAccount != nil {
		result.ExpiredAt = &result.VirtualAccount.ExpiryAt
	}

	if result.ChargePaymentMethodDetails != nil && result.Qr != nil {
		result.ExpiredAt = &result.Qr.ExpiryAt
	}

	response.SendApiResponseOK(w, result)
}
