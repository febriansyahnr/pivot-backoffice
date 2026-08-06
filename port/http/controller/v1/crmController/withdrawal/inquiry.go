package withdrawalCrmController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// InquiryTransaction	godoc
// @Summary				Inquiry withdrawal transaction status
// @Description			Endpoint for updating transaction status (inquiry from bank) when previous status was PENDING
// @ID					crm-withdrawal-inquiry-status
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param 				id		path		string true "Withdrawal ID"
// @Param 				Request	body		withdrawal.InquiryTransactionRequest true "Request body for inquiry transaction"
// @Success				200  	{object}	response.Response{data=withdrawal.InquiryTransactionResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/withdrawals/{id}/inquiry-transaction [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) InquiryTransaction(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/withdrawal/InquiryTransaction")
	defer segment.End()

	request := &withdrawal.InquiryTransactionRequest{
		Id: r.PathValue("id"),
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.service.InquiryTransaction(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
