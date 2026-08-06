package merchantCronHandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"go.uber.org/zap"
)

func (h *merchantCronHandler) DormantMerchant(ctx context.Context, dateStr string) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/merchant/DormantMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	err, start := error(nil), time.Now()
	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	var date time.Time
	if dateStr == "" {
		date = time.Now().In(tz)

	} else {
		if date, err = time.ParseInLocation(time.DateOnly, dateStr, tz); err != nil {
			h.logger.Error(ctx, "Invalid date format. Value: "+dateStr+" must be in YYYY-mm-dd format", zap.Error(err))
			return
		}
	}

	h.logger.Info(ctx, "Start to find and update dormant merchant")
	defer func() {
		duration := time.Now().UTC().Sub(start)

		h.logger.Info(ctx, "Dormant merchant completed", zap.Int64("duration", duration.Milliseconds()), zap.String("durationHuman", duration.String()), zap.Bool("completed", err == nil))
	}()

	if err = h.merchantSvc.DormantMerchant(ctx, date); err != nil {
		h.logger.Fatal(ctx, "An error occurred while dormant merchant", zap.Error(err))
	}
}
