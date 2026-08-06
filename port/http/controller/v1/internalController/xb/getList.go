package internalXbController

import (
	"context"
	errs "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalXbController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetList")
	defer segment.End()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// Set context expose unmapping request error
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	// Get query params
	var (
		page    int64 = 1
		perPage int64 = constant.DefaultPaginationPageSize
	)
	status := r.URL.Query().Get("status")
	uuid := r.URL.Query().Get("uuid")
	sortBy := r.URL.Query().Get("sortBy")
	sort := r.URL.Query().Get("sort")

	// Validation and parsing
	if err := c.bindOptionalInt64Query("page", r, &page); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if err := c.bindOptionalInt64Query("perPage", r, &perPage); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	startAt, endAt, err := c.getQueryForXbPayoutList(r)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	filter := &xbModel.GetPayoutFilterRequest{
		MerchantID: merchantID,
		UUID:       uuid,
		StartAt:    startAt,
		EndAt:      endAt,
		Status:     status,
		SortBy:     sortBy,
		Sort:       sort,
	}

	list, err := c.xbPayoutSvc.GetList(ctx, filter, page, perPage)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, list.Data, list.Meta)
}

func (c *InternalXbController) bindOptionalInt64Query(key string, r *http.Request, dst *int64) (err error) {
	val := r.URL.Query().Get(key)

	if val == "" {
		return

	} else if dst == nil {
		return errs.New("dst value can't be nil")
	}

	if *dst, err = strconv.ParseInt(val, 10, 64); err != nil {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid %s format. Use number format instead", key))
	}
	return nil
}

func (c *InternalXbController) getQueryForXbPayoutList(r *http.Request) (startAt, endAt time.Time, err error) {
	startDate := r.URL.Query().Get("startCreatedAt")
	endDate := r.URL.Query().Get("endCreatedAt")

	if startDate == "" || endDate == "" {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, errs.New("start date and end date must be filled"))
	}

	if startAt, err = time.Parse(constant.DatetimeRFC3339Format, startDate); err != nil {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidStartDateFmt)
	}

	if endAt, err = time.Parse(constant.DatetimeRFC3339Format, endDate); err != nil {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidEndDateFmt)
	}

	if startAt.After(endAt) {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrFilterDateInput)

	} else if (endAt.Sub(startAt).Hours() / 24) > 31 {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, errs.New("maximum date range 31 days"))
	}

	return startAt.UTC(), endAt.UTC(), nil
}
