package paymentMethodService

import (
	"context"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

// This function is used to anticipate future conditions where additional logic may be needed to retrieve the active payment methods for a payment request.
func (s *PaymentMethodService) GetActivePaymentMethodDetailForPaymentRequest(ctx context.Context, request paymentModel.GetActivePaymentMethodRequest) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/GetActivePaymentMethodDetailForPaymentRequest")
	defer segment.End()

	return s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, &request)
}
