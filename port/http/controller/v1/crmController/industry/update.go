package industry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMIndustryController) Update(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		request industryModel.UpdateIndustryRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/industry/Update")
	defer segment.End()

	uuid := chi.URLParam(r, "id")
	if uuid == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrIndustryIDRequired))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("error when decode body payload")))
		return
	}
	request.UUID = uuid

	if err := c.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.industryService.UpdateIndustry(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp.ToResponse())
}
