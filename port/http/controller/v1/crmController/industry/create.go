package industry

import (
	"encoding/json"
	"fmt"
	"net/http"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMIndustryController) Create(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		request industryModel.CreateIndustryRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/industry/Create")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("error when decode body payload")))
		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.industryService.CreateIndustry(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp.ToResponse())
}
