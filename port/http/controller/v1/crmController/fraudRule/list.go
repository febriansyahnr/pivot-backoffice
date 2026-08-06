package crmfraudrulecontroller

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMFraudRuleController) List(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/fraudRule/List")
	defer segment.End()
	var (
		page     int = constant.DefaultPage
		pageSize int = constant.DefaultPaginationPageSize
		err      error
	)

	var payload fraudrulesmodel.FraudRulesQuery
	payload.RuleName = r.URL.Query().Get("ruleName")
	payload.ReferenceType = r.URL.Query().Get("referenceType")

	strPage := r.URL.Query().Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := r.URL.Query().Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	if err := c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	fraudRules, err := c.fraudRuleService.List(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	list := []*fraudrulesmodel.FraudRulesResponse{}
	for _, config := range fraudRules.Data.([]*fraudrulesmodel.FraudRules) {
		list = append(list, config.ToResponse())
	}
	fraudRules.Data = list

	response.SendOpenApiResponseOK(w, fraudRules)
}
