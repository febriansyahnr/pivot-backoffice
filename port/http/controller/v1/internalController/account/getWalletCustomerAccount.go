package internalAccountController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetWalletCustomerAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/account/GetWalletCustomerAccount")
	defer segment.End()

	var (
		merchantId string
		payload    account_model.GetCustomerAccountRequest
	)
	httputil.BindMerchantID(r, &merchantId)
	payload.MerchantID = merchantId

	customerId := chi.URLParam(r, "customerId")
	payload.CustomerID = customerId

	account, err := h.accountSvc.GetWalletCustomerAccount(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, account.ToWalletResponse())
}

func (h *handler) CreateWalletCustomerAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/account/GetWalletCustomerAccount")
	defer segment.End()

	var (
		payload account_model.CreateCustomerAccountRequest
	)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	err := h.validator.Struct(payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	account, err := h.accountSvc.CreateAccount(ctx, &account_model.NewAccountRequest{
		ReferenceID: uuid.MustParse(payload.CustomerID),
		UserType:    constant.UserTypeCustomer,
		Usecase:     constant.TypeWallet,
		Currency:    constant.CurrencyIDR,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, account)

}
