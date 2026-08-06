package tnc

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (c *CRMTNCController) Activate(w http.ResponseWriter, r *http.Request) {
	c.setActive(w, r, true)
}

func (c *CRMTNCController) Deactivate(w http.ResponseWriter, r *http.Request) {
	c.setActive(w, r, false)
}

func (c *CRMTNCController) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/tnc/setActive")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	var (
		version *tncModel.TNCVersionResponse
		err     error
	)
	if active {
		version, err = c.service.ActivateTNCVersion(ctx, id)
	} else {
		version, err = c.service.DeactivateTNCVersion(ctx, id)
	}
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, version)
}
