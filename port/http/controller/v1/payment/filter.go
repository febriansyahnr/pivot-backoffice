package payment

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FilterPaymentHistory		godoc
// @Summary		Payment History Filter Endpoint
// @Description	This endpoint used to filter the payment history for internal operation
// @ID			api-payment-filter-dashboard
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=disbursementDashboardModel.DisbursementDashboardResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/ [get]
// @Security	Bearer
func (c *PaymentController) FilterPaymentHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/FilterPaymentHistory")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request, err := c.ParseFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate", "paymentStartDate", "paymentEndDate"); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	request.MerchantID = user.MerchantId
	result, err := c.paymentService.FilterPaymentHistory(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) ParseFilterParam(r *http.Request) (paymentModel.FilterPaymentHistoryOption, error) {
	var (
		opt paymentModel.FilterPaymentHistoryOption
		err error
	)
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartDate = d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndDate = d
	}

	if r.URL.Query().Get("paymentStartDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("paymentStartDate"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid paymentStartDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.PaymentStartDate = d
	}

	if r.URL.Query().Get("paymentEndDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("paymentEndDate"))
		if err != nil {
			return opt, errors.New(response.HttpErrRequest, fmt.Errorf("invalid paymentEndDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.PaymentEndDate = d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.Status = r.URL.Query().Get("status")
	opt.ReferenceID = r.URL.Query().Get("referenceId")
	opt.PaymentMethod = r.URL.Query().Get("paymentMethod")
	opt.SettlementModel = r.URL.Query().Get("settlementModel")

	return opt, nil
}
