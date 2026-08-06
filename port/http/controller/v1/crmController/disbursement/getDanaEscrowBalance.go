package crmDisbursementController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDanaEscrowBalance	godoc
// @Summary				get dana escrow balance
// @Description			get dana escrow balance
// @ID					crm-inquiry-escrow balance
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Success				200  	{object}	response.Response{data=routingProcessorModelEscrow.EscrowBalanceResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/balances/dana-balance [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (c *handler) GetDanaEscrowBalance(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "/port/http/controller/v1/crmController/disbursement/GetDanaEscrowBalance")
	defer span.End()

	balance, err := c.routingProcessorSvc.GetDanaEscrowBalance(ctx)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	if balance == nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
		return
	}

	response.SendGeneralResponseOK(w, balance)
}
