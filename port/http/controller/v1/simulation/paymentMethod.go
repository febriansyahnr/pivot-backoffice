package simulationController

import (
	"net/http"

	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *Handler) GetPaymentMethodForPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/simulation/GetPaymentMethodForPayment")
	defer segment.End()

	var (
		err                                            error
		payments                                       []*paymentModel.PaymentMethod
		virtualAccount, bankTransfer, creditCard, qris []*paymentModel.PaymentMethod
	)

	if payments, err = h.paymentMethodSvc.FindPaymentMethodByCategory(ctx, constant.PAYMENT_METHOD_CATEGORY_PAYMENT); err != nil {
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
