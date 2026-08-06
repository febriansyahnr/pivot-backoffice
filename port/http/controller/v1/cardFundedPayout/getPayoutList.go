package cardFundedPayoutController

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPayoutList godoc
// @Summary		Card Funded Payout list endpoint
// @Description	Get paginated list of card funded payouts
// @ID			api-card-funded-payout-list
// @Tags		API - Card Funded Payout
// @Accept		json
// @Produce		json
// @Param       startDate			query     	string  false  "filter by start date (ISO 8601)"
// @Param       endDate				query     	string  false  "filter by end date (ISO 8601)"
// @Param       transactionStatus	query     	string  false  "filter by transaction status (processing/success/failed)"
// @Param       approval			query     	string  false  "filter by approval status (approved/rejected/waiting)"
// @Param       searchId			query     	string  false  "search by payout ID or reference ID"
// @Param       page				query     	string  false  "pagination current page"
// @Param       perPage				query     	string  false  "items per page"
// @Param       sort				query     	string  false  "sort direction (ASC/DESC)"
// @Param       sortBy				query     	string  false  "field to sort by"
// @Success		200  				{object}	response.ApiResponse{data=[]cardFundedPayoutModel.GetPayoutListResponse,meta=commonModel.Meta}
// @Failure		500  				{object}	response.ApiResponse
// @Router		/api/v1/card-funded-payouts [get]
// @Security	Bearer
func (c *handler) GetPayoutList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetPayoutList")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Parse filter params
	filter, err := c.parseGetPayoutListFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Set merchant ID from token
	filter.MerchantID = user.MerchantId

	// Validate date range
	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	result, err := c.cardFundedPayoutService.GetPayoutList(ctx, filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *handler) parseGetPayoutListFilterParam(r *http.Request) (*cardFundedPayoutModel.FilterGetPayoutList, error) {
	var (
		filter cardFundedPayoutModel.FilterGetPayoutList
		err    error
	)

	// Set defaults
	filter.Page = 1
	filter.PerPage = constant.DefaultPaginationPageSize
	filter.Sort = "DESC"
	filter.SortBy = "createdAt"

	query := r.URL.Query()

	// Parse pagination
	if query.Get("page") != "" {
		filter.Page, err = strconv.ParseInt(query.Get("page"), 10, 64)
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if query.Get("perPage") != "" {
		filter.PerPage, err = strconv.ParseInt(query.Get("perPage"), 10, 64)
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	// Parse date filters
	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.StartCreatedAt = &d
	}

	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		filter.EndCreatedAt = &d
	}

	// Parse filters
	if query.Get("transactionStatus") != "" {
		filter.TransactionStatus = query.Get("transactionStatus")
	}

	if query.Get("approval") != "" {
		filter.ApprovalStatus = query.Get("approval")
	}

	if query.Get("searchId") != "" {
		filter.SearchID = query.Get("searchId")
	}

	if query.Get("sort") != "" {
		filter.Sort = query.Get("sort")
	}

	if query.Get("sortBy") != "" {
		filter.SortBy = query.Get("sortBy")
	}

	return &filter, nil
}
