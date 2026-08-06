package charges

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetChargeList		godoc
// @Summary		Get charges list Endpoint
// @Description	Get paginated list of charges
// @ID			api-charges-get-list
// @Tags		API - Charges
// @Accept		json
// @Produce		json
// @Param		page			query		int		false	"Page number (default: 1)"
// @Param		perPage			query		int		false	"Items per page (default: 10)"
// @Param		startDate		query		string	false	"Start date filter (YYYY-MM-DDTHH:mm:ssZ)"
// @Param		endDate			query		string	false	"End date filter (YYYY-MM-DDTHH:mm:ssZ)"
// @Param		sort			query		string	false	"Sort order: ASC or DESC (default: ASC)"
// @Param		sortBy			query		string	false	"Sort by field (default: createdAt)"
// @Param		id				query		string	false	"Filter by charge ID"
// @Param		status			query		string	false	"Filter by charge status"
// @Param		clientReferenceId	query	string	false	"Filter by client reference ID"
// @Param		paymentSessionId	query	string	false	"Filter by payment session ID"
// @Param		subMerchantId	query		string	false	"Sub-merchant ID"
// @Success		200  			{object}	response.ApiResponse{data=[]unifiedPaymentModel.ChargeResponse,meta=commonModel.Meta}
// @Failure		400  			{object}	response.ApiResponse
// @Failure		401  			{object}	response.ApiResponse
// @Failure		500  			{object}	response.ApiResponse
// @Router		/api/v1/charges [get]
// @Security	Bearer
func (c *ChargesController) GetChargeList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/charges/GetChargeList")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID := user.MerchantId
	if r.URL.Query().Get("subMerchantId") != "" {
		err := c.merchantService.ValidateSubMerchantParent(ctx, user.MerchantId, r.URL.Query().Get("subMerchantId"))
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}

		// Set parent merchant ID in context for sub-merchant requests
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, user.MerchantId)
		merchantID = r.URL.Query().Get("subMerchantId")
	}

	request, err := c.parseChargeFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate", "paymentStartDate", "paymentEndDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	request.MerchantID = merchantID
	result, err := c.unifiedPaymentService.GetChargeList(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *ChargesController) parseChargeFilterParam(r *http.Request) (unifiedPaymentModel.FilterChargeRequest, error) {
	var (
		opt unifiedPaymentModel.FilterChargeRequest
		err error
	)
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = d
	}

	if r.URL.Query().Get("paymentStartDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("paymentStartDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid paymentStartDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartPaymentDate = d
	}

	if r.URL.Query().Get("paymentEndDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("paymentEndDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid paymentEndDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndPaymentDate = d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.UUID = r.URL.Query().Get("id")
	opt.Status = r.URL.Query().Get("status")
	opt.ClientReferenceID = r.URL.Query().Get("clientReferenceId")
	opt.PaymentSessionID = r.URL.Query().Get("paymentSessionId")

	return opt, nil
}
