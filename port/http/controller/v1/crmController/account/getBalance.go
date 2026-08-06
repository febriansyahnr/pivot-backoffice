package account

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
	"net/http"
)

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/account/GetBalance")
	defer segment.End()

	// check if merchant id is valid
	merchantId, err := uuid.Parse(chi.URLParam(r, "merchantId"))
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}

	// TODO: Handle multi currency later.
	currency := r.URL.Query().Get("currency")
	if currency != "" && currency != constant.CurrencyIDR {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrAccountNotFound))
		return
	}

	usecaseType := r.URL.Query().Get("usecase")
	if usecaseType == "" {
		// default usecase is disbursement
		usecaseType = constant.TypeDisbursement
	}

	userTypeMerchant := r.URL.Query().Get("userType")
	if userTypeMerchant == "" {
		// default user type is merchant
		userTypeMerchant = constant.UserTypeMerchant
	}

	// Find account by usecase
	account, err := h.accountSvc.GetAccountByReferenceIDAndUsecase(ctx, merchantId, usecaseType, userTypeMerchant)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrDatabase, err))
		return
	} else if account == nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrNotFound, constant.ErrAccountNotFound))
		return
	}

	// Get available balance by account id
	availableBalance, err := h.orchestratorSvc.GetAvailableMerchantBalance(ctx, merchantId.String(), account.Name)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]any{
		"availableBalance": commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance).StringFixed(2),
		},
	})
}
