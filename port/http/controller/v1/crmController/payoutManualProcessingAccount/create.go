package payoutManualProcessingAccount

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (c *CRMPayoutManualProcessingAccountController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payoutManualProcessingAccount/Create")
	defer segment.End()

	var payload payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Validate merchantId is a UUID; existence is not checked (CRM is a trusted caller).
	if _, err := uuid.Parse(payload.MerchantID); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	account, err := c.service.Create(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, account)
}
