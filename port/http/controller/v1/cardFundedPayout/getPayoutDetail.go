package cardFundedPayoutController

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPayoutDetail godoc
// @Summary		Card Funded Payout detail endpoint
// @Description	Get detailed information of a card funded payout by ID
// @ID			api-card-funded-payout-detail
// @Tags		API - Card Funded Payout
// @Accept		json
// @Produce		json
// @Param 		payoutId		path		string true "Payout ID (UUID)"
// @Success		200  			{object}	response.ApiResponse{data=cardFundedPayoutModel.GetPayoutDetailResponse}
// @Failure		400  			{object}	response.ApiResponse
// @Failure		401  			{object}	response.ApiResponse
// @Failure		404  			{object}	response.ApiResponse
// @Failure		500  			{object}	response.ApiResponse
// @Router		/api/v1/card-funded-payouts/{payoutId} [get]
// @Security	Bearer
func (c *handler) GetPayoutDetail(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetPayoutDetail")
	defer span.End()

	// Get payout ID from URL
	payoutID := chi.URLParam(r, "payoutId")
	if err := uuid.Validate(payoutID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("payoutId is required and must be a valid UUID")))
		return
	}

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Build request
	request := &cardFundedPayoutModel.GetPayoutDetailRequest{
		PayoutID:   payoutID,
		MerchantID: user.MerchantId,
	}

	// Get payout detail from service
	result, err := c.cardFundedPayoutService.GetPayoutDetail(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
