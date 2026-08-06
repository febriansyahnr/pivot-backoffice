package internalPaymentMethodController

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

func (c *InternalPaymentMethodController) TopUpVAPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/paymentMethod/TopUpVAPaymentMethod")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	paymentMethodId := strings.TrimSpace(chi.URLParam(r, "paymentMethodId"))
	if paymentMethodId == "" {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrRequiredField("paymentMethodId"))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid payment method id")))
		return
	}

	disbursementTopUp, err := c.MerchantTopUpSvc.FindOrCreate(ctx, merchantId, constant.TypeDisbursement, paymentMethodId)
	if err != nil {
		if errors.Is(err, constant.ErrPaymentMethodNotFound) {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("payment method", paymentMethodId))
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, disbursementTopUp.ToResponse())
}
