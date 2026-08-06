package payment

import (
	"context"

	"go.uber.org/zap"
)

func (h *paymentCronHandler) PublishPendingPaymentExpirationEvent(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/payment/ExpirePendingPayment")
	defer segment.End()

	err := h.paymentService.PublishPaymentExpirationMessage(ctx)
	if err != nil {
		h.logger.Fatal(ctx, "Failed to publish payment expiration message", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "Successfully published payment expiration message")
}
