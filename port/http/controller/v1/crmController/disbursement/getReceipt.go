package crmDisbursementController

import (
	"encoding/json"
	"net/http"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		request = &disbursementModel.GetDisbursementReceiptCRMRequest{}
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/disbursement/GetReceipt")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	disbursementReceipt, err := h.disbursementSvc.GetReceiptByID(ctx, &disbursementModel.GetDisbursementReceiptRequest{
		ReferenceID: request.ReferenceID,
		MerchantID:  request.MerchantID,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, disbursementReceipt)
}
