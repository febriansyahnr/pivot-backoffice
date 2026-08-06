package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

var billingStatus = map[string]string{
	"":       "PENDING",
	"unpaid": "PENDING",
	"paid":   "SUCCESS",
}

func (h *CRMMerchantController) GetBillingFees(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/GetBillingFees")
	defer segment.End()

	query := r.URL.Query()

	request := merchant.BillingFeeRequest{
		MerchantId: r.PathValue("merchantId"),
		Status:     query.Get("status"),
		BillingDateRangeRequest: &merchant.BillingDateRangeRequest{
			StrStartDate: query.Get("startDate"),
			StrEndDate:   query.Get("endDate"),
		},
	}
	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	request.Status = billingStatus[request.Status]
	if err := request.ParseDateRangeRequestFromAsiaJakartaToUtc(); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.merchantSvc.GetBillingFees(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, result)
}

func (h *CRMMerchantController) PayBillingFees(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/PayBillingFees")
	defer segment.End()

	request := merchant.PayBillingFeeRequest{
		MerchantId: r.PathValue("merchantId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := request.ParseDateRangeRequestFromAsiaJakartaToUtc(); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.merchantSvc.PayBillingFees(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, result)
}
