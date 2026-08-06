package internalPaymentMethodController

import (
	"context"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalPaymentMethodController) GetVAPaymentMethods(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/paymentMethod/GetVAPaymentMethods")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	request := &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantId,
		Category:   paymentConst.PAYMENT_METHOD_CATEGORY_MERCHANT_TOPUP,
		Type:       paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
	}
	data, err := c.PaymentMethodSvc.GetPaymentMethodByMerchant(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	responseData := []*paymentModel.PlatformPaymentMethodResponse{}
	for _, item := range data {
		responseData = append(responseData, item.ToPlatformResponseModel())
	}

	response.SendOpenApiResponseOK(w, responseData)
}
