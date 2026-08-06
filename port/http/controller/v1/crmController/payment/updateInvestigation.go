package v1CrmPaymentController

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) UpdateInvestigation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/UpdateInvestigation")
	defer segment.End()

	paymentID := r.PathValue("id")
	if err := uuid.Validate(paymentID); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	var request paymentModel.UpdateInvestigationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.paymentSvc.UpdateInvestigationStatus(ctx, paymentID, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
