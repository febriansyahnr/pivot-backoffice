package processorCallbackController

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	flipProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/flipProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// FlipInquiryAccountCallback is a http handler for Flip Inquiry Account Callback.
// This endpoint accept form data with key "data" which is a json string of
// flipProcessorModel.AccountInquiryRequest. The request will be decoded and
// validated. If the request is valid, the handler will send a success response
// with a message "success". Otherwise, it will send an error response with
// the appropriate error message.
// this function will used when was implementing service
func (h *processorCallbackController) FlipInquiryAccountCallback(w http.ResponseWriter, r *http.Request) {
	var (
		ctx                 = r.Context()
		flipRequestCallback = &flipProcessorModel.AccountInquiryRequest{}
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
	u, err := url.ParseQuery("data=" + data)
	if err != nil {
		h.logger.Error(ctx, "Error parsing query string:", logger.Error(err))
		return
	}

	if err := json.NewDecoder(strings.NewReader(u.Get("data"))).Decode(&flipRequestCallback); err != nil {
		h.logger.Error(ctx, err.Error())
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrValidation))
		return
	}

	callbackRequest := flipRequestCallback.ToAccountInquiryResponse()

	err = h.routingProcessorSvc.ProcessAccountInquiryCallback(ctx, callbackRequest)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]any{
		"message": "success",
	})
}
