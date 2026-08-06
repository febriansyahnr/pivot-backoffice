package tnc

import (
	"encoding/json"
	"net/http"

	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMTNCController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/tnc/Create")
	defer segment.End()

	var payload tncModel.CreateTNCVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	version, err := c.service.CreateTNCVersion(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, version)
}
