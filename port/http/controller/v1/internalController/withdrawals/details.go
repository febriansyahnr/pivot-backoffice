package internalWithdrawalsController

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (h *InternalWithdrawalController) GetWithdrawalByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/withdrawals/GetWithdrawalByID")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantAuth.MerchantId

	// Bind sub-merchant id when id is sent via header
	httputil.BindSubmerchantID(r, &merchantId)

	id := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(id); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("withdrawalId"))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("invalid withdrawal id format")))
		return
	}

	result, err := h.withdrawalSvc.GetById(ctx, &withdrawal.WithdrawalDetailRequest{
		Id:          id,
		AccountName: constant.AccountNamePayment,
		MerchantId:  merchantId,
	})
	if err != nil {
		if errors.Is(err, constant.ErrDataNotFound) {
			err = constant.NewErrResourceNotFound("withdrawal detail", id)
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result.ToOpenAPIWithdrawalResponse())
}
