package industry

import (
	"net/http"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMIndustryController) GetAll(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		request = industryModel.SearchIndustryRequest{}
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/industry/GetAll")
	defer segment.End()

	queryParams := r.URL.Query()
	request.Keyword = queryParams.Get("keyword")

	resp, err := c.industryService.GetAllIndustries(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	data := []*industryModel.IndustryResponse{}
	for _, industry := range resp {
		data = append(data, industry.ToResponse())
	}

	response.SendApiResponseOK(w, data)

}
