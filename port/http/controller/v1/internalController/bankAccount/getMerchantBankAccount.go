package internalBankAccountController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalBankAccountController) GetMerchantBankAccount(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/bankAccount/GetMerchantBankAccount")
	defer segment.End()

	requesterMerchant := r.Header.Get(constant.HeaderXMerchantId)
	merchantId := chi.URLParam(r, "merchantId")

	bankAccount, err := c.svc.GetByMerchantID(ctx, &bankAccount.GetMerchantBankAccountRequest{
		MerchantID:          merchantId,
		RequesterMerchantID: requesterMerchant,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, bankAccount.ToResponse())
}
