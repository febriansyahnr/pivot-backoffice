package merchantHandler

import "context"

func (h *handler) MigrateMerchantSecretsToEncryption(ctx context.Context) error {
	ctx, span := otelTracer.Start(ctx, "port/console/handler/merchant/MigrateMerchantSecretsToEncryption")
	defer span.End()

	return h.service.MigrateMerchantSecretsToEncryption(ctx)
}
