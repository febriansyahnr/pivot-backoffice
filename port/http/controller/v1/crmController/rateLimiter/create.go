package crmRateLimiterController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	rateLimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMRateLimiterController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/rateLimiter/Create")
	defer segment.End()

	merchantID := chi.URLParam(r, "merchantId")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrMerchantIdIsRequired))
		return
	}

	var payload rateLimiterModel.CreateRateLimitConfiguration
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	payload.MerchantID = merchantID
	if err := c.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	config, err := c.svc.Create(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, config)
}
