package feeHandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"go.uber.org/zap"
)

func (h *feeHandler) DetermineFeeTierLvlFromMonthlyTPV(ctx context.Context, dateStr string) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/fee/DetermineFeeTierLvlFromMonhtlyTPV")
	defer segment.End()

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

	h.logger.Info(ctx, "Start determine fee tier level from monthly TPV period "+date.Format("2006-01"))
	defer func() {
		duration := time.Since(start)

		h.logger.Info(
			ctx, "Determine fee tier level from monthly TPV completed",
			zap.Int64("duration", duration.Milliseconds()), zap.String("durationHuman", duration.String()), zap.Bool("completed", err == nil),
		)
	}()

	if err = h.service.DetermineFeeTierLvlFromMonthlyTPV(ctx, date); err != nil {
		h.logger.Fatal(ctx, "An error occurred while determine fee tier level from monthly tpv", zap.Error(err))
	}
}
