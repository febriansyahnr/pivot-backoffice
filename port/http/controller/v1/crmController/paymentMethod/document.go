package crmPaymentMethodController

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"net/http"
)

func (h *handler) GetRequiredMerchantDocumentList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/paymentMethod/GetRequiredMerchantDocumentList")
	defer span.End()

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	paymentMethodID := chi.URLParam(r, "paymentMethodId")
	if err := uuid.Validate(paymentMethodID); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, errors.New("paymentMethodId is required")))
		return
	}

	resp, err := h.paymentMethodSvc.GetRequiredMerchantDocuments(ctx, &paymentMethodModel.GetRequiredMerchantDocumentsRequest{
		MerchantID:      merchantID,
		PaymentMethodID: paymentMethodID,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}
