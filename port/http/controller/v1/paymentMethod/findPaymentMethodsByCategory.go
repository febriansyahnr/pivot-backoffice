package paymentMethodController

import (
	"fmt"
	"net/http"
	"strings"

	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

// FindPaymentMethodByCategory	godoc
// @Summary						List of payment methods by category
// @Description					List of payment methods by category
// @ID							api-payment-method-list-by-category
// @Tags						API - Payment Method
// @Accept						json
// @Produce						json
// @Param						category	query	string	true	"Category of payment method"
// @Success						200  	{object}	response.ApiResponse{data=[]paymentModel.PaymentMethodResponse}
// @Failure						500  	{object}	response.ApiResponse
// @Router						/api/v1/payment-methods/:category [get]
func (c *PaymentMethodController) FindPaymentMethodByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/paymentMethod/FindPaymentMethodByCategory")
	defer segment.End()

	var (
		err                                            error
		payments                                       []*paymentModel.PaymentMethod
		virtualAccount, bankTransfer, creditCard, qris []*paymentModel.PaymentMethod
	)

	cat := chi.URLParam(r, "category")
	if cat == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("category is required")))
		return
	}

	if payments, err = c.paymentMethodSvc.FindPaymentMethodByCategory(r.Context(), strings.ToUpper(cat)); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	for _, payment := range payments {
		switch payment.Type {
		case constant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
			virtualAccount = append(virtualAccount, payment)
		case constant.PAYMENT_METHOD_BANK_TRANSFER:
			bankTransfer = append(bankTransfer, payment)
		case constant.PAYMENT_METHOD_CREDIT_CARD:
			creditCard = append(creditCard, payment)
		case constant.PAYMENT_METHOD_QRIS:
			qris = append(qris, payment)
		}
	}

	response.SendApiResponseOK(w, &paymentModel.PaymentMethodResponseGroup{
		VirtualAccount: toListPaymentMethodsResponse(virtualAccount),
		BankTransfer:   toListPaymentMethodsResponse(bankTransfer),
		CreditCard:     toListPaymentMethodsResponse(creditCard),
		QRIS:           toListPaymentMethodsResponse(qris),
	})
}

func toListPaymentMethodsResponse(payments []*paymentModel.PaymentMethod) []*paymentModel.PaymentMethodResponse {
	res := make([]*paymentModel.PaymentMethodResponse, 0)

	if len(payments) == 0 {
		return res
	}

	for _, payment := range payments {
		res = append(res, payment.ToResponse())
	}
	return res
}
