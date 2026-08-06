package creditcard

import (
	"encoding/json"
	"net/http"
	"time"

	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GeneratePaymentToken		godoc
// @Summary		Generate payment token for internal use
// @Description	Generate a new payment token with specified expiry time
// @ID			api-creditcard-extend-payment-token-internal
// @Tags		API - CreditCard Internal
// @Accept		json
// @Produce		json
// @Param		request	body		GeneratePaymentTokenRequest	true	"Generate payment token request"
// @Success		200  	{object}	GeneratePaymentTokenResponse
// @Failure		400  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/internal/cards/extend-payment-token [post]
// @Security	InternalAPIKey
func (c *Controller) GeneratePaymentToken(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/creditcard/GeneratePaymentToken")
	defer segment.End()

	var req unifiedPaymentModel.GeneratePaymentTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Error(ctx, "error decoding generate payment token request", logger.Error(err))
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err := c.validate.Struct(req); err != nil {
		c.logger.Error(ctx, "error validating generate payment token request", logger.Error(err))
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Parse expiry time
	expiryAt, err := time.Parse(time.RFC3339, req.ExpiryAt)
	if err != nil {
		c.logger.Error(ctx, "error parsing expiry time", logger.Error(err))
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.paymentSvc.HandleStrictExpiry(ctx, req.PaymentID); err != nil {
		c.logger.Error(ctx, "error handling strict expiry", logger.Error(err))
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Generate payment token
	token, err := c.paymentSvc.GeneratePaymentToken(ctx, req.PaymentID, expiryAt)
	if err != nil {
		c.logger.Error(ctx, "error generating payment token", logger.Error(err))
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, unifiedPaymentModel.GeneratePaymentTokenResponse{
		Token: token,
	})
}
