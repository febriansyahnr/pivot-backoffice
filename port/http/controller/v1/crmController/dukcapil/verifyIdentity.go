package dukcapil

import (
	"encoding/json"
	"net/http"

	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// VerifyIdentity		godoc
// @Summary				Dukcapil Identity Verification
// @Description			Use Dukcapil to verify identity information
// @ID					crm-dukcapil-verify-identity
// @Tags				CRM - Dukcapil
// @Accept				json
// @Produce				json
// @Param				merchantId	query	string							false	"Merchant ID for storing results"
// @Param 				request 	body 	dukcapilmodel.VerifyRequest 	true 	"Identity verification request"
// @Success				200  		{object}	response.ApiResponse{data=dukcapilmodel.IdentityVerificationResponse}
// @Failure				400  		{object}	response.ApiResponse
// @Failure				500  		{object}	response.ApiResponse
// @Router				/crm/v1/dukcapil/verify-identity [post]
// @Header       		all {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMDukcapilController) VerifyIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/dukcapil/VerifyIdentity")
	defer segment.End()

	queryParams := r.URL.Query()
	merchantID := queryParams.Get("merchantId")

	var verifyRequest dukcapilmodel.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&verifyRequest); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(verifyRequest); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	request := &dukcapilmodel.IdentityVerificationRequest{
		MerchantID:    merchantID,
		VerifyRequest: &verifyRequest,
	}

	result, err := c.dukcapilService.VerifyIdentity(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
