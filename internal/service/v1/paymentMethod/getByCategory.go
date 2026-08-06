package paymentMethodService

import (
	"context"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) FindPaymentMethodByCategory(
	ctx context.Context, category string) ([]*paymentModel.PaymentMethod, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/FindPaymentMethodByCategory")
	defer segment.End()

	// Get Payment Method by category
	paymentMethod, err := s.paymentMethodRepo.GetAllPaymentMethodByCategory(ctx, category)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method data by category", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	return paymentMethod, nil
}
