package cardFundedPayoutController

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetTransactionConfig(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/GetTransactionConfig")
	defer span.End()

	var (
		query            = r.URL.Query()
		amount           = 0.0
		settlementMethod = constant.SettlementMethodStandard
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if query.Get("settlementMethod") != "" &&
		slices.Contains([]string{constant.SettlementMethodInstant, constant.SettlementMethodStandard}, query.Get("settlementMethod")) {

		settlementMethod = query.Get("settlementMethod")
	}

	if query.Get("amount") != "" {
		amount, _ = strconv.ParseFloat(query.Get("amount"), 64)
	}

	ctx, configs, err := h.merchantService.GetMerchantIdForConfigs(ctx, user.MerchantId, false)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	var feeDetail feeModel.FeeResponseder
	if configs.MerchantType == constant.MerchantTypeSubMerchant {
		parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string)

		feeRequest := &feeModel.GetTrxFeeOnBehalfRequest{
			MerchantId:        parentMerchantId,
			SubMerchantId:     user.MerchantId,
			Reference:         constant.ReferencePaymentFundedPayout,
			PaymentMethod:     paymentConst.PAYMENT_METHOD_CREDIT_CARD,
			TransactionAmount: amount,
		}
		feeDetail, err = h.feeService.GetTransactionFeeOnBehalf(ctx, feeRequest)

	} else {
		feeRequest := &feeModel.GetFeeRequest{
			MerchantID:       user.MerchantId,
			Reference:        constant.ReferencePaymentFundedPayout,
			PaymentMethod:    paymentConst.PAYMENT_METHOD_CREDIT_CARD,
			ReferenceAmount:  amount,
			SettlementMethod: settlementMethod,
		}
		_, feeDetail, err = h.feeService.GetFeeCalculationAndDetail(ctx, feeRequest)

	}
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, cardFundedPayoutModel.TransactionConfigResponse{
		FeeDetail: feeDetail,
	})
}
