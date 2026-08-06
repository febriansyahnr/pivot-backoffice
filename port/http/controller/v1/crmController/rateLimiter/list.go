package crmRateLimiterController

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	rateLimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMRateLimiterController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/ratelimiter/GetList")
	defer segment.End()
	var (
		page     int = constant.DefaultPage
		pageSize int = constant.DefaultPaginationPageSize
		err      error
	)

	merchantID := chi.URLParam(r, "merchantId")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrMerchantIdIsRequired))
		return
	}

	var payload rateLimiterModel.MerchantRateLimitRequest
	payload.MerchantID = merchantID
	payload.Status = r.URL.Query().Get("status")

	strPage := r.URL.Query().Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendOpenApiResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := r.URL.Query().Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendOpenApiResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	if err := c.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	configs, err := c.svc.List(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, configs)
}
