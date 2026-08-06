package withdrawalController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	withdrawalModel "github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetInsights		godoc
// @Summary		Withdrawal insight Endpoint
// @Description	Withdrawal Insight Endpoint, currently we provide withdrawal balance
// @ID			api-withdrawal-insight
// @Tags		API - withdrawal
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=withdrawalModel.withdrawalInsightResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/withdrawals/insights [get]
// @Security	Bearer
func (h *handler) GetInsights(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawals/GetInsights")
	defer segment.End()
	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	insight, err := h.service.GetTodayWithdrawalInsight(ctx, withdrawalModel.WithdrawalInsightRequest{
		MerchantID: user.MerchantId,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, insight)
}
