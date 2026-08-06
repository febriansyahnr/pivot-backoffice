package submerchant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantInternalController) GetSubMerchantBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/api/v1/subMerchant/GetBalance")
	defer segment.End()

	var err error

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchantInfo := r.Context().Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	loggedInMerchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &loggedInMerchantId)
	loggedInUserType := constant.UserTypeMerchant
	httputil.BindLoggedInUserType(r, &loggedInUserType)

	subMerchantId := chi.URLParam(r, "id")
	if subMerchantId == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrRequiredSubmerchantId))
		return
	}

	// Treat parent mismatch as "not found" for consistency with detail endpoint
	// and to avoid leaking existence of sub-merchants owned by other parents.
	if err = c.merchantSvc.ValidateSubMerchantParent(ctx, merchant.MerchantId, subMerchantId); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("sub-account", subMerchantId))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrSubMerchantNotFound))
		return
	}

	// TODO: Handle multi currency later.
	currency := r.URL.Query().Get("currency")
	if currency != "" && currency != constant.CurrencyIDR {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("currency"))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("unsupported currency, only IDR is supported")))
		return
	}

	usecaseType := r.URL.Query().Get("usecase")
	if usecaseType == "" {
		// default usecase is disbursement
		usecaseType = constant.TypeDisbursement
	}

	subMerchantUUID, _ := uuid.Parse(subMerchantId)
	// Find account by usecase
	account, err := c.accountSvc.GetAccountByReferenceIDAndUsecase(ctx, subMerchantUUID, usecaseType, constant.UserTypeMerchant)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrDatabase, err))
		return
	} else if account == nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("account", subMerchantId))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrAccountNotFound))
		return
	}

	// Get available balance by account id
	availableBalance, err := c.orchestratorSvc.GetAvailableMerchantBalance(ctx, subMerchantId, account.Name)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	// Get pending balance by account id
	pendingBalance, err := c.orchestratorSvc.GetPendingBalance(ctx, subMerchantId, account.Name)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, map[string]any{
		"availableBalance": commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance).StringFixed(2),
		},
		"pendingBalance": commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(pendingBalance).StringFixed(2),
		},
		"totalBalance": commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance + pendingBalance).StringFixed(2),
		},
	})
}
