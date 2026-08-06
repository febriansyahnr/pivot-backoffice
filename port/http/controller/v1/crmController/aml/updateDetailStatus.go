package aml

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UpdateDetailStatus		godoc
// @Summary			Update AML screening detail status
// @Description		Update the status of a specific screening detail item by profileID
// @ID				crm-aml-update-detail-status
// @Tags			CRM - AML Screening
// @Accept			json
// @Produce			json
// @Param			profileId	path	string	true	"Profile ID of the detail to update"
// @Param			merchantId	query	string	true	"Merchant ID"
// @Param			request		body	amlcommon.UpdateDetailStatusRequest	true	"Status update request"
// @Success			200  	{object}	response.ApiResponse
// @Failure			400  	{object}	response.ApiResponse
// @Failure			404  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/crm/v1/aml/screening/{profileId}/status [put]
// @Header       	all {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMAmlController) UpdateDetailStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/aml/UpdateDetailStatus")
	defer segment.End()

	queryParams := r.URL.Query()
	merchantID := queryParams.Get("merchantId")
	if merchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDRequired))
		return
	}

	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	var payload amlcommon.UpdateDetailStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	err := c.amlService.UpdateDetailStatusByProfileId(ctx, profileID, merchantID, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	response.SendApiResponseOK(w, map[string]string{
		"message": "Status updated successfully",
	})
}
