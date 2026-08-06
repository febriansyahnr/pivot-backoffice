package account

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (b *account) CalculateAllMerchantEODBalance(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/account/CalculateAllMerchantEODBalance")
	defer segment.End()

	err := b.accountSvc.CalculateAccountEodBalance(ctx)
	if err != nil {
		b.logger.Fatal(ctx, "error when calculate eod balance", zap.Error(err))
	}
}

func (b *account) CalculateAllMerchantDailyAccountTransaction(ctx context.Context, location *time.Location) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/account/CalculateAllMerchantDailyAccountTransaction")
	defer segment.End()

	err := b.accountSvc.CalculateDailyAccountTransaction(ctx, location)
	if err != nil {
		b.logger.Fatal(ctx, "error when calculate daily account transaction", zap.Error(err))
	}
}
