package crmfraudrulecontroller

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
)

func (c *CRMFraudRuleController) Create(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		payload fraudrulesmodel.CreateFraudRuleRequest
	)
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/fraudRule/Create")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	// validate weight between 0 and 1
	if payload.Weight.LessThan(decimal.Zero) || payload.Weight.GreaterThan(decimal.NewFromInt(1)) {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrFraudRuleWeight))
		return
	}

	fraudRule, err := c.fraudRuleService.Create(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, fraudRule)
}
