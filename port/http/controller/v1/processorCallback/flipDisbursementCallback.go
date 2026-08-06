package processorCallbackController

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	flipProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/flipProcessor/bankTransfer"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *processorCallbackController) FlipDisbursementCallback(w http.ResponseWriter, r *http.Request) {
	var (
		ctx                 = r.Context()
		flipRequestCallback = &flipProcessorModel.BankTransferResponse{}
		callbackRequest     = &routingProcessorModel.BankTransferResponseData{}
	)
	if err := r.ParseForm(); err != nil {
		h.logger.Error(ctx, "error parsing form", logger.Error(err))
		response.SendApiResponseError(ctx, w, errors.New("failed to parse form"))
		return
	}

	ctx, span := otelTracer.Start(ctx, "internal/http/controller/v1/processorCallback/FlipDisbursementCallback")
	defer span.End()

	// decode body from x-request-From
	data := r.FormValue("data")
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&flipRequestCallback); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrValidation))
		return
	}

	// mapping to payment callback request
	callbackRequest = flipRequestCallback.ToBankTransferResponse()
	callbackRequest.ProcessorReference = constant.FlipPGProcessor

	// process do callback disbursement
	err := h.disbursementSvc.ProcessUpdateTransferStatus(ctx, callbackRequest)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// process transaction disbursement
	response.SendApiResponseOK(w, map[string]any{
		"message": "success",
	})
}
