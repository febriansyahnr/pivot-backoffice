package xbPayoutController

import (
	"errors"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *xbPayoutController) GetXbPayoutDashboardInsights(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetXbPayoutDashboardInsights")
	defer segment.End()

	userTokenClaim, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "insightStartDate", "insightEndDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	queryParams := r.URL.Query()

	request := disbursementModel.GetXbPayoutDashboardInsightRequest{
		MerchantId: userTokenClaim.MerchantId,
	}
	request.StartDate, _ = time.Parse(time.RFC3339Nano, queryParams.Get("insightStartDate"))
	request.EndDate, _ = time.Parse(time.RFC3339Nano, queryParams.Get("insightEndDate"))

	// Validation of value formats, time ranges, and other rules are already handled in ValidateReportDateRangeFromRequest.
	if request.StartDate.IsZero() || request.EndDate.IsZero() {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("start date or end date cannot be empty")))
		return
	}

	result, err := h.xbPayoutSvc.GetXbPayoutDashboardInsights(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrInternal, constant.ErrInternalServerForUser))
		return
	}
	response.SendApiResponseOK(w, result)
}
