package disbursementController

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDisbursementInsight godoc
// @Summary Get disbursement insights for homepage dashboard
// @Description Get comprehensive disbursement insights including waiting, delayed, status breakdown, and failure reasons
// @Tags disbursement-insights
// @Accept json
// @Produce json
// @Param insightStartDate query string false "Start date for insights (RFC3339 format)"
// @Param insightEndDate query string false "End date for insights (RFC3339 format)"
// @Param includePreviousPeriod query bool false "Include previous period comparison data (default: true)"
// @Security BearerAuth
// @Success 200 {object} response.ApiResponse{data=disbursementModel.DisbursementInsightResponse}
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /api/v1/disbursement-insights [get]
func (c *Controller) GetDisbursementInsight(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetDisbursementInsight")
	defer segment.End()

	var (
		err error
		req = disbursementModel.GetDisbursementInsightFilter{
			InsightStartDate:      util.GetCurrentDateOfLocation(util.GetJakartaTimeLocation()).UTC(),
			InsightEndDate:        util.GetCurrentDateOfLocation(util.GetJakartaTimeLocation()).AddDate(0, 0, 1).UTC(),
			IncludePreviousPeriod: true, // Default to true
		}
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	req.MerchantID = user.MerchantId

	date, exist := httputil.GetDateTimeQueryParam(r, "insightStartDate", time.RFC3339)
	if exist {
		req.InsightStartDate = date
	}

	date, exist = httputil.GetDateTimeQueryParam(r, "insightEndDate", time.RFC3339)
	if exist {
		req.InsightEndDate = date
	}

	if includePreviousPeriodStr, exist := httputil.GetQueryParam(r, "includePreviousPeriod"); exist {
		if includePreviousPeriodStr == "false" {
			req.IncludePreviousPeriod = false
		}
	}

	insights, err := c.disbursementSvc.GetDisbursementInsight(ctx, req)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, insights)
}
