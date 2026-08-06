package paymentMethodService

import (
	"context"

	"github.com/google/uuid"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) Create(ctx context.Context, payload *paymentMethodModel.CreatePaymentMethodRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/Create")
	defer segment.End()

	uuidV7, _ := uuid.NewV7()
	payload.UUID = uuidV7.String()
	err := s.paymentMethodRepo.CreatePaymentMethod(ctx, payload)
	if err != nil {
		s.logger.Error(ctx, "failed to upsert payment method", logger.Error(err), logger.Any("payload", payload))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}
