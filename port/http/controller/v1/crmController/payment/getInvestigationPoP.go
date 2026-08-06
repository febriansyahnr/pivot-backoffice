package v1CrmPaymentController

import (
	"errors"
	"net/http"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (h *handler) GetInvestigationProofOfPayment(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(
		r.Context(),
		"port/http/controller/v1/crmController/payments/GetInvestigationProofOfPayment",
	)
	defer segment.End()

	request := model.GetInvestigationProofOfPaymentRequest{
		PaymentID: r.PathValue("paymentId"),
	}
	if err := uuid.Validate(request.PaymentID); err != nil {
		response.SendGeneralResponseError(
			w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid payment ID format")),
		)
		return
	}

	result, err := h.paymentSvc.GetInvestigationProofOfPayment(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}
