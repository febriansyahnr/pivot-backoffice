package paymentService

import (
	"context"
	"fmt"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

func (s *PaymentService) GetImages(ctx context.Context) (paymentModel.ImageResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetImages")
	defer segment.End()

	data, err := s.paymentRepo.RetrieveImages(ctx)
	if err != nil {
		return paymentModel.ImageResponse{}, fmt.Errorf("failed to retrieve images: %w", err)
	}

	return data, nil
}
