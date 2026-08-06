package paymentService

import (
	"context"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

func (s *PaymentService) changePaymentMethod(ctx context.Context, request *paymentModel.UpdateUnifiedPaymentRequest, payment *paymentModel.Payment) (*paymentModel.UpdateUnifiedPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/changePaymentMethod")
	defer segment.End()

	if request.Customer.Name == "" && payment.CustomerID != "" {
		// select customer data from existing payment
		customer, err := s.customerRepo.GetCustomerById(ctx, payment.CustomerID, payment.MerchantID)
		if err != nil {
			return nil, err
		}

		request.Customer = paymentModel.PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customer.FirstName + " " + customer.LastName,
			Email:      customer.Email,
			Phone:      customer.PhoneNumber,
		}
	}

	paymentRequest := &paymentModel.CreateUnifiedPaymentRequest{
		PaymentID:            payment.UUID,
		MerchantID:           payment.MerchantID,
		ClientReferenceID:    request.ClientReferenceID,
		Amount:               *request.Amount,
		PaymentMethod:        request.PaymentMethod,
		Customer:             request.Customer,
		ExpiryAt:             *request.ExpiryAt,
		PaymentMethodOptions: request.PaymentMethodOptions,
	}

	result, err := s.internal.CreateUnifiedPayment(ctx, paymentRequest)
	if err != nil {
		return nil, err
	}
	return &paymentModel.UpdateUnifiedPaymentResponse{
		ID:                result.ID,
		ClientReferenceID: result.ClientReferenceID,
		Amount:            result.Amount,
		PaymentMethod:     result.PaymentMethod,
		ExpiryAt:          result.ExpiryAt,
		PaymentLink:       result.PaymentLink,
	}, nil
}
