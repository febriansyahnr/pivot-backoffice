package creditcard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/shopspring/decimal"
)

// CreatePayment	godoc
// @Summary			Create credit card payment
// @Description		Create credit card payment
// @ID				api-credit-card-create-payment
// @Tags			API - Credit Card
// @Accept			json
// @Produce			json
// @Param			Request	body		card.CreateCardPaymentRequest true "JSON Body for Create Credit Card Payment Request"
// @Success			200  	{object}	response.OpenApiResponse
// @Failure			500  	{object}	response.OpenApiResponse
// @Router			/api/v1/cards/payment-session [post]
// @Security		Bearer
func (c *Controller) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/creditcard/CreatePayment")
	defer segment.End()

	var (
		err       error
		minAmount = decimal.NewFromFloat(constant.CreditCardMinAmount)

		now = time.Now()
	)

	var requestPayload creditcardModel.CreateCardPaymentRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if requestPayload.Amount.Cmp(minAmount) < 0 {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest,
			constant.ErrCreditcardMinAmount))
		return
	}

	merchantToken, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	// set additional request
	requestPayload.MerchantID, err = uuid.Parse(merchantToken.MerchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest,
			constant.ErrCreditcardInvalidUUID))
		return
	}

	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantToken.MerchantId)

		merchantToken.MerchantId = subMerchantId
	}

	switch requestPayload.AuthenticationMethod {
	case constant.CreditCardMethodChallenge, constant.CreditCardMethodFrictionless:
	default:
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest,
			constant.ErrCreditcardInvalidAuthenticationMethod))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer monitor.WriteAndSend(
		ctx, "creditcard-create-payment", now, ww, err, func() []string {
			return []string{
				fmt.Sprintf("reference_id:%s", requestPayload.ReferenceID),
				fmt.Sprintf("bank_merchant_id:%s", requestPayload.BankMerchantID),
				fmt.Sprintf("amount:%s", requestPayload.Amount),
				fmt.Sprintf("currency:%s", requestPayload.Currency),
				fmt.Sprintf("authentication_method:%s", requestPayload.AuthenticationMethod),
				fmt.Sprintf("merchant_id:%s", requestPayload.MerchantID),
			}
		},
	)

	data, err := c.creditcardSvc.CreatePayment(ctx, requestPayload)
	if err != nil {
		response.SendOpenApiResponseError(ww, err)
		return
	}
	response.SendOpenApiResponseOK(ww, data)
}
