package v1CrmPaymentController

import (
	"encoding/json"
	"net/http"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	var (
		request = &paymentModel.GetPaymentReceiptCRMRequest{}
	)

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payment/GetReceipt")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	paymentReceipt, err := h.paymentSvc.GetReceiptByID(ctx, &paymentModel.GetPaymentReceiptRequest{
		ReferenceID: request.ReferenceID,
		MerchantID:  request.MerchantID,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+paymentReceipt.Filename+"\"")
	// w.Header().Set("Content-Length", strconv.Itoa(len(paymentReceipt.PDF)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(paymentReceipt.PDF)
}
