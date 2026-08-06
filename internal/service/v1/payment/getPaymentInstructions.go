package paymentService

import (
	"context"
	"fmt"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentService) GetPaymentInstructions(ctx context.Context, paymentMethod string) ([]paymentModel.InstructionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentInstructions")
	defer segment.End()

	instructionsList, err := s.paymentRepo.RetrieveInstructions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment instructions: %w", err)
	}

	if paymentMethod != "" {
		for _, instructions := range instructionsList {
			if instructions.Title == paymentMethod {
				return []paymentModel.InstructionResponse{instructions}, nil
			}
		}

		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, fmt.Errorf("instructions for payment method %s not found", paymentMethod))
	}
	return instructionsList, nil
}
