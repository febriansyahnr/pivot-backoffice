package payment

import (
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PaymentController) GetInvestigationSummary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetInvestigationSummary")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Parse required date range
	startDateStr := r.URL.Query().Get("startDate")
	if startDateStr == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("startDate is required")))
		return
	}

	endDateStr := r.URL.Query().Get("endDate")
	if endDateStr == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("endDate is required")))
		return
	}

	startDate, err := time.Parse(util.UTCLayout, startDateStr)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
		return
	}

	endDate, err := time.Parse(util.UTCLayout, endDateStr)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	opt := paymentModel.GetInvestigationSummaryOption{
		MerchantID: user.MerchantId,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	result, err := c.paymentService.GetInvestigationSummary(ctx, opt)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
