package tnc

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (c *CRMTNCController) List(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/tnc/List")
	defer segment.End()

	var (
		page     = constant.DefaultPage
		pageSize = constant.DefaultPaginationPageSize
		err      error
	)

	query := r.URL.Query()
	var payload tncModel.TNCVersionQuery

	payload.Version = query.Get("version")
	payload.Title = query.Get("title")
	if isActiveStr := query.Get("isActive"); isActiveStr != "" {
		isActive := isActiveStr == "true" || isActiveStr == "1"
		payload.IsActive = &isActive
	}
	payload.SortBy = query.Get("sortBy")
	payload.Sort = query.Get("sort")

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

	result, err := c.service.ListTNCVersions(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, result)
}

func (c *CRMTNCController) Detail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/tnc/Detail")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	version, err := c.service.GetTNCVersion(ctx, id)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, version)
}
