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

func (c *CRMRateLimiterController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/rateLimiter/Update")
	defer segment.End()

	merchantID := chi.URLParam(r, "merchantId")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrMerchantIdIsRequired))
		return
	}

	id := chi.URLParam(r, "id")
	_, err := uuid.Parse(id)
	if err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	var payload rateLimiterModel.UpdateRateLimitConfiguration
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	payload.ID = id
	payload.MerchantID = merchantID
	if err := c.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	config, err := c.svc.Update(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, config)

}
