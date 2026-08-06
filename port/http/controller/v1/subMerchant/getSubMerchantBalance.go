package subMerchant

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *SubMerchantController) GetSubMerchantBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/api/v1/subMerchant/GetBalance")
	defer segment.End()

	merchant, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	subMerchantId := chi.URLParam(r, "id")
	if subMerchantId == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrRequiredSubmerchantId))
		return
	}

	err := h.merchantSvc.ValidateSubMerchantParent(ctx, merchant.MerchantId, subMerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrIncorrectSubMerchant))
		return
	}

	// TODO: Handle multi currency later.
	currency := r.URL.Query().Get("currency")
	if currency != "" && currency != constant.CurrencyIDR {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrAccountNotFound))
		return
	}

	usecaseType := r.URL.Query().Get("usecase")
	if usecaseType == "" {
		// default usecase is disbursement
		usecaseType = constant.TypeDisbursement
	}

	merchantBalance, err := h.orchestratorSvc.GetMerchantBalance(ctx, orchestrator_model.GetMerchantBalanceRequest{
		Date:        time.Now().UTC(),
		MerchantID:  subMerchantId,
		BalanceName: usecaseType,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, map[string]any{
		"availableBalance": merchantBalance.AvailableBalance,
		"pendingBalance":   merchantBalance.PendingBalance,
		"totalBalance":     merchantBalance.TotalBalance,
	})
}
