package adjustment

import (
	"net/http"

	adjustmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetHoldedMerchantBalance	godoc
// @Summary					Get holded merchant balance
// @Description				Get holded merchant balance
// @ID						crm-balance-hold-get
// @Tags					API - CRM
// @Accept					json
// @Produce					json
// @Param					merchantId	query		string	true "Merchant ID"
// @Param					accountType	query		string	true "Account Type"	Enums(PAYMENT, VIRTUAL_TERMINAL, WALLET)
// @Success					200  		{object}	response.Response{data=adjustment.GetHoldedMerchantBalanceResponse}
// @Failure					500  		{object}	response.Response
// @Router					/crm/v1/balances/hold [get]
// @Header       			all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetHoldedMerchantBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/balance/GetHoldedMerchantBalance")
	defer segment.End()

	request := adjustmentModel.GetHoldedMerchantBalanceRequest{
		MerchantId:  r.URL.Query().Get("merchantId"),
		AccountType: r.URL.Query().Get("accountType"),
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.service.GetHoldedMerchantBalance(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, result)
}
