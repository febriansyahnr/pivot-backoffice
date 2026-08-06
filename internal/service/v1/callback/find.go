package callbackService

import (
	"context"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"

	"github.com/google/uuid"
)

func (s *CallbackService) FindCallbackByMerchantIdAndCallbackName(ctx context.Context, merchantId uuid.UUID, callbackName string) (*callbackModel.Callback, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/callback/FindCallbackByMerchantIdAndCallbackName")
	defer span.End()

	return s.callbackRepo.FindCallbackByNameAndMerchantID(ctx, callbackName, merchantId)
}
