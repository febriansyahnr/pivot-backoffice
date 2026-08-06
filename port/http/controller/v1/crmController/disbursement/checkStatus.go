package crmDisbursementController

import (
	"encoding/json"
	"net/http"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) CheckTransactionStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "/port/http/controller/v1/crmController/disbursement/CheckTransactionStatus")
	defer span.End()

	var (
		req  disbursementModel.CheckDisbursementTransactionStatusRequest
		resp []*disbursementModel.CheckDisbursementTransactionStatusResponse
	)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	err = h.validator.Struct(req)
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	resp, err = h.disbursementSvc.CheckTransactionStatus(ctx, &req)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
