package crmCreditcardController

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetTransactionList		godoc
// @Summary				Get creditcard transaction List
// @Description			Get creditcard transaction List
// @ID					get-creditcard-transaction-list
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				Request	body		card.GetTransactionListRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=card.VoidRequest}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/creditcard/transaction/list [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetTransactionList(w http.ResponseWriter, r *http.Request) {

	var (
		ctx = r.Context()

		err error
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/creditcard/GetTransactionList")
	defer segment.End()

	query := r.URL.Query()

	request := &creditcardModel.GetTransactionListRequest{
		DateFrom:            query.Get("dateFrom"),
		DateTo:              query.Get("dateTo"),
		TrxType:             query.Get("type"),
		ChargeStatus:        query.Get("chargeStatus"),
		VoidStatus:          query.Get("voidStatus"),
		ClientTransactionID: query.Get("clientTransactionId"),
		IssuingBank:         query.Get("issuingBank"),
		CardFingerprint:     query.Get("cardFingerprint"),
		PaymentUUID:         query.Get("paymentUuid"),
		MerchantID:          query.Get("merchantId"),
		ChargeFrom:          query.Get("chargeFrom"),
		ChargeTo:            query.Get("chargeTo"),
		RefundFrom:          query.Get("refundFrom"),
		RefundTo:            query.Get("refundTo"),
	}

	page := query.Get("page")
	if page != "" {
		request.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		request.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	transactions, err := h.creditcardSvc.GetTransactionList(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, transactions)
}
