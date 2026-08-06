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
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetInvestigationList godoc
// @Summary		Investigation List Endpoint
// @Description	Get list of payments under investigation
// @ID			api-cases-list
// @Tags		API - Cases Management
// @Accept		json
// @Produce		json
// @Param		paymentReferenceId	query	string	false	"Payment Reference ID (alias: paymentReferenceID)"
// @Param		investigationStatus	query	string	false	"Investigation Status (INVESTIGATION_IN_PROCESS, INVESTIGATION_SUCCESS, INVESTIGATION_FAILED)"
// @Param		paymentMethod		query	string	false	"Payment Method Type (VIRTUAL_ACCOUNT, QRIS, etc)"
// @Param		channel				query	string	false	"Payment Channel/Bank Name"
// @Param		fromDate			query	string	false	"From Date (YYYY-MM-DDTHH:mm:ssZ)"
// @Param		toDate				query	string	false	"To Date (YYYY-MM-DDTHH:mm:ssZ)"
// @Param		page				query	int		false	"Page number"
// @Param		perPage				query	int		false	"Items per page (alias: limit)"
// @Param		sortBy				query	string	false	"Sort by field"
// @Param		sort				query	string	false	"Sort direction (ASC, DESC)"
// @Success		200  	{object}	response.ApiResponse
// @Failure		400  	{object}	response.ApiResponse
// @Failure		401  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/cases [get]
// @Security	Bearer
func (c *PaymentController) GetInvestigationList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payment/GetInvestigationList")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request, err := c.parseInvestigationFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Set merchant ID from token
	request.MerchantID = user.MerchantId

	result, err := c.paymentService.GetInvestigatedPayments(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) parseInvestigationFilterParam(r *http.Request) (paymentModel.GetInvestigatedPaymentsFilterRequest, error) {
	var (
		filter paymentModel.GetInvestigatedPaymentsFilterRequest
		err    error
	)

	// Support both paymentReferenceID and paymentReferenceId for backward compatibility
	filter.PaymentReferenceID = r.URL.Query().Get("paymentReferenceID")
	if filter.PaymentReferenceID == "" {
		filter.PaymentReferenceID = r.URL.Query().Get("paymentReferenceId")
	}

	filter.InvestigationStatus = r.URL.Query().Get("investigationStatus")
	filter.PaymentMethod = r.URL.Query().Get("paymentMethod")
	filter.Channel = r.URL.Query().Get("channel")

	if r.URL.Query().Get("fromDate") != "" {
		fromDate, err := time.Parse(util.UTCLayout, r.URL.Query().Get("fromDate"))
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid fromDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.FromDate = &fromDate
	}

	if r.URL.Query().Get("toDate") != "" {
		toDate, err := time.Parse(util.UTCLayout, r.URL.Query().Get("toDate"))
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid toDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.ToDate = &toDate
	}

	if r.URL.Query().Get("page") != "" {
		filter.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead"))
		}
	}

	limitParam := r.URL.Query().Get("limit")
	if limitParam == "" {
		limitParam = r.URL.Query().Get("perPage")
	}
	if limitParam != "" {
		filter.Limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid limit format. Use number format instead"))
		}
	}

	filter.SortBy = r.URL.Query().Get("sortBy")
	filter.Sort = r.URL.Query().Get("sort")

	return filter, nil
}
