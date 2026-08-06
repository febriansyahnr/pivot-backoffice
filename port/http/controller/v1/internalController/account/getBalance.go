package internalAccountController

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/account/GetBalance")
	defer segment.End()

	var (
		merchantId string
		err        error
	)
	merchantClaim, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId = merchantClaim.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)
	merchantUUID, _ := uuid.Parse(merchantId)

	// TODO: Handle multi currency later.
	currency := r.URL.Query().Get("currency")
	if currency != "" && currency != constant.CurrencyIDR {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("unsupported currency, only IDR is supported")))
		return
	}

	usecaseType := r.URL.Query().Get("usecase")
	if usecaseType == "" {
		// default usecase is disbursement
		usecaseType = constant.TypeDisbursement
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameAccount,
			MetricName:           constant.MetricNameAccountBalance,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          merchantClaim.MerchantId,
				"usecase":             usecaseType,
				"onBehalfSubmerchant": merchantClaim.MerchantId != merchantId,
			},
		}
		errMetric := customMetric.RecordCustomMetric(ctx, metricData)
		if errMetric != nil {
			h.logger.Error(ctx, "failed to record get balance custom metric", logger.Error(errMetric))
		}
	}()

	// Find account by usecase
	account, err := h.accountSvc.GetAccountByReferenceIDAndUsecase(ctx, merchantUUID, usecaseType, constant.UserTypeMerchant)
	if err != nil {
		err = pkgErrors.New(response.HttpErrDatabase, err)
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrDatabase, err))
		return
	} else if account == nil {
		err = pkgErrors.New(response.HttpErrNotFound, constant.ErrAccountNotFound)
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	// Get available balance by account id
	availableBalance, err := h.orchestratorSvc.GetAvailableMerchantBalance(ctx, merchantId, account.Name)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, map[string]any{
		"availableBalance": commonModel.Amount{
			Currency: account.Currency,
			Value:    decimal.NewFromFloat(availableBalance).StringFixed(2),
		},
	})
}
