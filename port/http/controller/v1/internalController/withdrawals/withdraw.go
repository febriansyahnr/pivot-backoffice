package internalWithdrawalsController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalWithdrawalController) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/withdrawals/Withdraw")
	defer segment.End()

	var (
		payload                      withdrawal.OpenAPIWithdrawalRequest
		merchantId, parentMerchantId string
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	parentMerchantId = merchantAuth.MerchantId
	merchantId = merchantAuth.MerchantId

	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantId = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	payload.MerchantId = merchantId
	payload.ParentMerchantId = parentMerchantId
	if payload.IsFullAmount && payload.Amount.Value == "" {
		payload.Amount = commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    "0",
		}
	}
	if err := c.validate.Struct(payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	request := payload.ToWithdrawalRequest()

	// When a withdrawal is triggered through the Open API,
	// it is necessary to record who initiated the process in order to distinguish whether it was performed
	// directly or on behalf (specifically for the platform).
	request.UserId = merchantAuth.MerchantId

	resp, err := c.withdrawalSvc.Create(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp.ToOpenAPIWithdrawalResponse(&payload))
}
