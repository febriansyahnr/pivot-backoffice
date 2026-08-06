package cardFundedPayoutController

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *handler) GetVendorList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetVendorList")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Parse merchantId from user token
	merchantID, err := uuid.Parse(user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	var (
		page     int = constant.DefaultPage
		pageSize int = constant.DefaultPaginationPageSize
	)

	query := r.URL.Query()

	var payload vendorModel.VendorQuery
	payload.MerchantID = merchantID

	payload.Name = query.Get("name")
	payload.Status = query.Get("status")
	payload.SortBy = query.Get("sortBy")
	payload.Sort = query.Get("sort")

	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		payload.StartDate = &d
	}

	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		payload.EndDate = &d
	}

	strPage := query.Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := query.Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	result, err := c.vendorService.List(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *handler) GetVendorDetail(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetVendorDetail")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	_, err := uuid.Parse(id)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	vendor, err := c.vendorService.Detail(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Validate that the vendor belongs to the user's merchant
	if vendor.MerchantID != user.MerchantId {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrNotFound, constant.ErrVendorNotFound))
		return
	}

	response.SendApiResponseOK(w, vendor.ToResponse())
}
