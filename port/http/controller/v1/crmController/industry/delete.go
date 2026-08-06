package industry

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMIndustryController) Delete(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/industry/Delete")
	defer segment.End()

	uuid := chi.URLParam(r, "id")
	if uuid == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrIndustryIDRequired))
		return
	}

	err := c.industryService.DeleteIndustry(ctx, uuid)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
