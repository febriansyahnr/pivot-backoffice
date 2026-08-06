package disbursementController

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDisbursementDashboard		godoc
// @Summary		Disbursement dashboard endpoint
// @Description	Disbursement dashboard endpoint
// @ID			api-disbursement-dashboard
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=disbursementDashboardModel.DisbursementDashboardResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/disbursements/dashboard [get]
// @Security	Bearer
func (c *Controller) GetDisbursementDashboard(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetDisbursementDashboard")
	defer segment.End()

	var (
		err error
		req = disbursementDashboardModel.GetDisbursementDashboardFilter{
			InsightStartDate: util.GetCurrentDateOfLocation(util.GetJakartaTimeLocation()).UTC(),
			InsightEndDate:   util.GetCurrentDateOfLocation(util.GetJakartaTimeLocation()).AddDate(0, 0, 1).UTC(),
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

	dashboard, err := c.disbursementDashboardSvc.Get(ctx, req)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, dashboard)
}
