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

func (c *PaymentController) GetVCCTerminalList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payments/GetVCCTerminalList")
	defer segment.End()

	userInfo, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request, err := parseGetVCCTerminalListFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Set merchant ID from token
	request.MerchantID = userInfo.MerchantId

	result, err := c.paymentService.GetVCCTerminalList(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func parseGetVCCTerminalListFilterParam(r *http.Request) (paymentModel.GetVCCTerminalListFilterRequest, error) {
	var (
		filter paymentModel.GetVCCTerminalListFilterRequest
		err    error
	)

	filter.Status = r.URL.Query().Get("status")
	filter.ChargeID = r.URL.Query().Get("chargeId")
	filter.ReferenceID = r.URL.Query().Get("referenceId")

	if r.URL.Query().Get("chargeStartDate") != "" {
		fromDate, err := time.Parse(util.UTCLayout, r.URL.Query().Get("chargeStartDate"))
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid chargeStartDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.ChargeStartDate = fromDate
	}

	if r.URL.Query().Get("chargeEndDate") != "" {
		fromDate, err := time.Parse(util.UTCLayout, r.URL.Query().Get("chargeEndDate"))
		if err != nil {
			return filter, errors.New(response.HttpErrRequest, fmt.Errorf("invalid chargeEndDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.ChargeEndDate = fromDate
	}

	// Parse page and limit
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
