package crmRateLimiterController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMRateLimiterController) Detail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/rateLimiter/Detail")
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

	config, err := c.svc.Detail(ctx, merchantID, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, config)

}
