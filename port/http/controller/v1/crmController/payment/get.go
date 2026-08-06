package v1CrmPaymentController

import (
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"net/http"
)

func (h *handler) GetDetailByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/GetDetailByID")
	defer segment.End()

	var (
		err error
	)

	paymentID := r.PathValue("id")
	if err = uuid.Validate(paymentID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	result, err := h.paymentSvc.GetDetailByID(ctx, paymentID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}

func (h *handler) GetSplitRoutingByTransferID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/GetSplitRoutingByTransferID")
	defer segment.End()

	var (
		err error
	)

	paymentID := r.PathValue("paymentId")
	if err = uuid.Validate(paymentID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	transferID := r.PathValue("transferId")
	if err = uuid.Validate(transferID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrSplitRoutingPaymentNotValid))
		return
	}

	result, err := h.paymentSvc.GetSplitRoutingByTransferID(ctx, paymentID, transferID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
