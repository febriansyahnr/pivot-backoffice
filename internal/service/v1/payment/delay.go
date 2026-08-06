package paymentService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

// getDelayedConfigDuration returns the delay duration for processing a payment based on its method and the retry count from the context.
// If the payment is nil or its method is not eligible for delayed processing (not credit card or e-wallet), it returns 0.
// If the retry count is not present in the context or exceeds the available configuration, it returns 0.
// Otherwise, it returns the configured delay duration for the current retry count.
func (s *PaymentService) getDelayedConfigDuration(ctx context.Context, payment *paymentModel.Payment) (int, time.Duration) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetDelayedConfig")
	defer span.End()

	delayCfg := s.config.UnifiedPaymentConfig.ExpiringProcessedBackoffMinutes

	if payment == nil ||
		(payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_CREDIT_CARD &&
			payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_EWALLET) {
		s.logger.Info(ctx, "payment method is not have delayed config for processing payment")
		return 0, 0
	}

	retryCount, ok := ctx.Value(constant.CtxRabbitMQRetryCount).(int32)
	if !ok || retryCount < 0 {
		s.logger.Info(ctx, "the context did not have retry count value")
		return 0, 0
	}

	if int(retryCount) >= len(delayCfg) {
		return int(retryCount), -1
	}

	return int(retryCount), time.Duration(delayCfg[int(retryCount)]) * time.Minute
}
