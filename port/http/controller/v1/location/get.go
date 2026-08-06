package location

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Get				godoc
// @Summary			Endpoint to obtain address data such as province, city and district
// @Description		Endpoint to obtain address data such as province, city and district
// @ID				address-locations
// @Tags			API - Address Location
// @Accept			json
// @Produce			json
// @Param			name			path	string	true	"Name of location such as province, city and district"
// @Param        	provinceId    	query	string  false	"The city list inquiry through the province ID"
// @Param        	cityId	    	query	string  false	"The district list inquiry through the city ID"
// @Success			200  	{object}	response.Response{data=location.LocationResp}
// @Failure			500  	{object}	response.Response
// @Router			/crm/v1/address/locations/{name} [get]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/UploadDocument")
	defer segment.End()

	request := &location.LocationReq{
		Name:       r.PathValue("name"),
		ProvinceId: r.URL.Query().Get("provinceId"),
		CityId:     r.URL.Query().Get("cityId"),
	}
	if err := h.validate.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.Get(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
