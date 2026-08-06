package country

import (
	"net/http"
	"strings"

	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMCountryController) GetAll(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		request = countryModel.SearchFilterRequest{}
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/country/GetAll")
	defer segment.End()

	queryParams := r.URL.Query()
	language := queryParams.Get("lang")
	if strings.ToLower(language) == "id" {
		request.NameID = queryParams.Get("name")
	} else {
		request.Name = queryParams.Get("name")
	}

	resp, err := c.countryService.GetAll(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	data := []*countryModel.CountryResponse{}
	for _, country := range resp {
		data = append(data, country.ToResponse())
	}

	response.SendApiResponseOK(w, data)

}
