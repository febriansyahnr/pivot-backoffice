package crmDisbursementController

import (
	"encoding/json"
	"net/http"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetPayoutStatusAndRouting	godoc
// @Summary				Get payout status and routing history for CRM (Single Transaction)
// @Description			Get detailed payout status including bank routing history and real-time status from processors for a single transaction. Use /batch-payout-status for multiple transactions.
// @ID					crm-get-payout-status-and-routing
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param 				Request	body		disbursementModel.CRMSinglePayoutStatusRequest true "Payout status request with single referenceId"
// @Success				200  	{object}	response.Response{data=disbursementModel.CRMPayoutStatusResponse}
// @Failure				400  	{object}	response.Response
// @Failure				404  	{object}	response.Response
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/disbursements/payout-status [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetPayoutStatusAndRouting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/disbursement/GetPayoutStatusAndRouting")
	defer segment.End()

	request := &disbursementModel.CRMSinglePayoutStatusRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	result, err := h.disbursementSvc.GetPayoutStatusAndRouting(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}

// GetBatchPayoutStatusAndRouting	godoc
// @Summary				Get multiple payout status and routing history for CRM
// @Description			Get detailed payout status for multiple transactions including bank routing history and real-time status from processors. Returns batch response with individual results and summary.
// @ID					crm-get-batch-payout-status-and-routing
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param 				Request	body		disbursementModel.CRMBatchPayoutStatusRequest true "Batch payout status request - must include referenceIds array"
// @Success				200  	{object}	response.Response{data=disbursementModel.CRMBatchPayoutStatusResponse}
// @Failure				400  	{object}	response.Response
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/disbursements/batch-payout-status [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetBatchPayoutStatusAndRouting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/disbursement/GetBatchPayoutStatusAndRouting")
	defer segment.End()

	request := &disbursementModel.CRMBatchPayoutStatusRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrValidation, err))
		return
	}

	result, err := h.disbursementSvc.GetBatchPayoutStatusAndRouting(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
