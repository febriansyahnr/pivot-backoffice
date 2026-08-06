package v1InternalUnifiedPaymentController

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) FindPaymentByReferenceId(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/FindPaymentByReferenceId")
	defer segment.End()

	var (
		err     error
		payment *paymentModel.UnifiedPaymentResponse
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	id := chi.URLParam(r, "referenceId")
	if id == "" {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))

		return
	}

	if payment, err = c.paymentSvc.GetPaymentByReferenceId(r.Context(), id, merchantAuth.MerchantId); err != nil {
		response.SendOpenApiResponseError(w, err)

		return
	}

	response.SendOpenApiResponseOK(w, payment)

}
