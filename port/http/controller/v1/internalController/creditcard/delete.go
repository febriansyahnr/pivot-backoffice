package creditcard

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RemoveCardByCustomerIDAndTokenID is a function used to remove card tokenization from a customer
func (c *Controller) RemoveCardByCustomerIDAndTokenID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/creditcard/RemoveCardByCustomerIDAndTokenID")
	defer segment.End()

	request := unifiedPaymentModel.RemoveCardTokenizationRequest{
		MerchantID: r.PathValue("merchantId"),
		CustomerID: r.PathValue("customerId"),
		TokenID:    r.PathValue("tokenId"),
	}
	request.PaymentID, _ = ctx.Value(constant.CtxPaymentID).(string)

	if err := c.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	if err := c.creditcardSvc.RemoveCardTokenization(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, unifiedPaymentModel.RemoveCardTokenizationResponse{
		CustomerID: request.CustomerID,
		TokenID:    request.TokenID,
	})
}
