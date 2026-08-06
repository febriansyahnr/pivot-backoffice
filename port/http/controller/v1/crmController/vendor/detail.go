package vendor

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMVendorController) Detail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/vendor/Detail")
	defer segment.End()

	id := chi.URLParam(r, "id")
	_, err := uuid.Parse(id)
	if err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	vendor, err := c.vendorService.Detail(ctx, id)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, vendor.ToResponse())
}
