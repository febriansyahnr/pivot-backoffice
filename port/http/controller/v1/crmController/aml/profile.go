package aml

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Screening		godoc
// @Summary			AML Profile
// @Description		Use Profile to check for AML
// @ID				crm-aml-profile
// @Tags			CRM - AML Profile
// @Accept			json
// @Produce			json
// @Param			provider	query	string	true	"provider of AML"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			crm/v1/aml/profile [post]
// @Header       	all {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMAmlController) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/aml/Screening")
	defer segment.End()

	queryParams := r.URL.Query()
	provider := queryParams.Get("provider")
	if provider == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrProviderRequired))
		return
	}

	// allow merchant id to be empty
	// in case no need to save the result in merchant
	merchantID := queryParams.Get("merchantId")
	profileID := queryParams.Get("profileId")

	var payload amlcommon.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.amlService.Profile(ctx, &payload, provider, merchantID, profileID)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	response.SendApiResponseOK(w, resp)
}
