package v1CrmPaymentController

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) InquiryByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/InquiryByID")
	defer segment.End()

	var (
		err error
	)

	paymentID := r.PathValue("id")
	if err = uuid.Validate(paymentID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	result, err := h.paymentSvc.InquiryPayment(ctx, &paymentModel.InquiryPaymentRequest{
		PaymentID: paymentID,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
