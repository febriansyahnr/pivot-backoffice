package processorCallbackController

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	snapUtil "github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
)

func (h *processorCallbackController) DanaDisbursementCallback(w http.ResponseWriter, r *http.Request) {
	var (
		ctx             = r.Context()
		callbackRequest = &routingProcessorModel.BankTransferResponseData{}
		rawBody         map[string]any
		err             error
	)

	ctx, span := otelTracer.Start(ctx, "controller/v1/processorCallback/DanaDisbursementCallback")
	defer span.End()

	// Decode to raw body
	if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.HttpErrValidation))
		return
	}

	externalID := r.Header.Get("X-EXTERNAL-ID")
	latestTransactionStatus, _ := rawBody["latestTransactionStatus"].(string)

	// Marshal then unmarshal again for the callback request
	jsonBody, _ := json.Marshal(rawBody)
	if err := json.Unmarshal(jsonBody, &callbackRequest); err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.HttpErrValidation))
		return
	}

	// Map the required field
	callbackRequest.ExternalID = externalID
	callbackRequest.ProcessorReference = constant.DanaPGProcessor
	callbackRequest.Status = snapUtil.GetStatusFromLatestTransactionCode(latestTransactionStatus)

	// Update the status
	err = h.disbursementSvc.ProcessUpdateTransferStatus(ctx, callbackRequest)
	if err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiSnapResponseOK(ctx, w, map[string]any{
		"responseCode":    "2004300",
		"responseMessage": "Successful",
	})
}
