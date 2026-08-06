package cardFundedPayoutController

import (
	"encoding/json"
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

func (c *handler) CreateSavedCard(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/CreateSavedCard")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	payload := cardFundedPayoutModel.CreateSavedCardRequest{
		MerchantID: user.MerchantId,
		CreatedBy:  user.UUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.cardFundedPayoutService.CreateSavedCard(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}

func (c *handler) GetSavedCardList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetSavedCardList")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Filter
	request, err := c.parseCardFundedPayoutFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Add Merchant to Filter
	request.MerchantID = user.MerchantId

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	result, err := c.cardFundedPayoutService.GetSavedCardList(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *handler) parseCardFundedPayoutFilterParam(r *http.Request) (cardFundedPayoutModel.FilterGetSavedCardList, error) {
	var (
		opt cardFundedPayoutModel.FilterGetSavedCardList
		err error
	)
	opt.Page = 1
	opt.PerPage = 1000
	opt.Sort = "DESC"
	opt.SortBy = "createdAt"

	query := r.URL.Query()

	if query.Get("page") != "" {
		opt.Page, err = strconv.Atoi(query.Get("page"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if query.Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(query.Get("perPage"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = &d
	}

	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = &d
	}

	if query.Get("sort") != "" {
		opt.Sort = query.Get("sort")
	}

	if query.Get("sortBy") != "" {
		opt.SortBy = query.Get("sortBy")
	}

	return opt, nil
}
