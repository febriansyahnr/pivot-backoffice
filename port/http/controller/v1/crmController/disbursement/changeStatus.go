package crmDisbursementController

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ChangeStatus			godoc
// @Summary				change disbursement status
// @Description			this endpoint used to change disbursement transaction status and consume by internal tool
// @ID					crm-change-disbursement-transaction-status
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				Request	body		disbursementModel.ChangeDisbursementTransactionStatus true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=disbursementModel.DisbursementChangeTransactionStatusResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/disbursements/change-transaction-status [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "/port/http/controller/v1/crmController/disbursement/ChangeStatus")
	defer span.End()

	var (
		req  disbursementModel.ChangeDisbursementTransactionStatusRequest
		resp []disbursementModel.ChangeDisbursementTransactionStatusResponse
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

	// When referenceNumber is provided, exactly one disbursementId must be supplied
	// since the bank reference number maps to a single disbursement record.
	if strings.TrimSpace(req.ReferenceNumber) != "" && len(req.DisbursementIDS) != 1 {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, errors.New("when referenceNumber is provided, exactly one disbursementId is allowed")))
		return
	}

	resp = h.disbursementSvc.ChangeDisbursementTransactionStatus(ctx, req)

	response.SendApiResponseOK(w, resp)
}
