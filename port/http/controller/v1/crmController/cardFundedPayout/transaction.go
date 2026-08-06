package crmCardFundedPayoutController

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetPayoutTransactionList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/cardFundedPayout/GetPayoutTransactionList")
	defer segment.End()

	err := error(nil)
	query := r.URL.Query()

	request := model.GetPayoutTransactionListRequest{
		MerchantID:    query.Get("merchantId"),
		StrStartDate:  query.Get("startDate"),
		StrEndDate:    query.Get("endDate"),
		TrxStatus:     query.Get("trxStatus"),
		TrxReasonType: query.Get("trxReasonType"),
	}
	if err = h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	request.StartDate, _ = time.ParseInLocation(time.RFC3339, request.StrStartDate, time.UTC)
	request.EndDate, _ = time.ParseInLocation(time.RFC3339, request.StrEndDate, time.UTC)

	if request.StartDate.After(request.EndDate) {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidDateRange))
		return
	}

	transactions, err := h.service.GetPayoutTransactionList(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, transactions)
}

func (h *handler) PatchPayoutTransactionStatus(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/cardFundedPayout/PatchPayoutTransactionStatus")
	defer segment.End()

	request := model.PatchPayoutTransactionStatusRequest{
		PayoutID: r.PathValue("payoutId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.service.UpdatePayoutTransactionStatus(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, result)
}
