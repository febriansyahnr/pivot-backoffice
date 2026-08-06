package cardFundedPayoutController

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPayoutInsights godoc
// @Summary		Card Funded Payout insights endpoint
// @Description	Get total amount and total transaction for card funded payouts with WAITING approval status
// @ID			api-card-funded-payout-insights
// @Tags		API - Card Funded Payout
// @Accept		json
// @Produce		json
// @Param       startDate	query     	string  false  "filter by start date (ISO 8601)"
// @Param       endDate		query     	string  false  "filter by end date (ISO 8601)"
// @Success		200  {object}	response.ApiResponse{data=model.GetPayoutInsightsResponse}
// @Failure		500  {object}	response.ApiResponse
// @Router		/api/v1/card-funded-payouts/insights [get]
// @Security	Bearer
func (h *handler) GetPayoutInsights(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetPayoutInsights")
	defer span.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	filter := &model.FilterGetPayoutInsights{
		MerchantID: user.MerchantId,
	}

	query := r.URL.Query()
	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		filter.StartCreatedAt = &d
	}
	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		filter.EndCreatedAt = &d
	}

	result, err := h.cardFundedPayoutService.GetPayoutInsights(ctx, filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
