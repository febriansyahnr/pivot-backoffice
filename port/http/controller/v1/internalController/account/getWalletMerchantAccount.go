package internalAccountController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetWalletMerchantAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/account/GetWalletMerchantAccount")
	defer segment.End()

	parentMerchantId := uuid.MustParse(r.Header.Get(constant.HeaderXMerchantId))
	merchantId := chi.URLParam(r, "merchantId")
	merchantUUID, err := uuid.Parse(merchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}

	account, err := h.accountSvc.GetWalletMerchantAccount(ctx, parentMerchantId, merchantUUID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, account.ToWalletResponse())
}
