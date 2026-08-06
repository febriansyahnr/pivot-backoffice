package vendor

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMVendorController) List(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/vendor/List")
	defer segment.End()

	var (
		page     int = constant.DefaultPage
		pageSize int = constant.DefaultPaginationPageSize
		err      error
	)

	query := r.URL.Query()

	var payload vendorModel.VendorQuery

	// Parse merchantId as UUID if provided
	merchantIDStr := query.Get("merchantId")
	if merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
			return
		}
		payload.MerchantID = merchantID
	}

	payload.Name = query.Get("name")
	payload.Status = query.Get("status")
	payload.SortBy = query.Get("sortBy")
	payload.Sort = query.Get("sort")

	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, stderrors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		payload.StartDate = &d
	}

	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, stderrors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		payload.EndDate = &d
	}

	strPage := query.Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := query.Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	if err := c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	result, err := c.vendorService.List(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, result)
}
