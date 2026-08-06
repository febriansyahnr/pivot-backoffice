package feeHandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"go.uber.org/zap"
)

func (h *feeHandler) DeductBalanceForIndirectFeeType(ctx context.Context, dateStr string) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/fee/DeductBalanceForIndirectFeeType")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	var (
		err   = error(nil)
		start = time.Now().UTC()
	)

	h.logger.Info(ctx, "Start balance deduction for indirect fee type")
	defer func() {
		duration := time.Now().UTC().Sub(start)

		h.logger.Info(
			ctx, "Balance deduction completed",
			zap.Int64("duration", duration.Milliseconds()), zap.String("durationHuman", duration.String()), zap.Bool("completed", err == nil),
		)
	}()

	// Expected value in Asia/Jakarta time zone
	// - dateStr
	// - date
	var date time.Time
	if dateStr == "" {
		date = time.Now().In(tz)

	} else {
		if date, err = time.ParseInLocation(time.DateTime, dateStr, tz); err != nil {
			h.logger.Error(ctx, "Invalid date format. Value: "+dateStr+" must be in YYYY-mm-dd HH:MM:SS format", zap.Error(err))
			return
		}
	}

	if err = h.service.DeductBalanceForIndirectFeeType(ctx, date); err != nil {
		h.logger.Fatal(ctx, "An error occurred while deduct balance for indirect fee type", zap.Error(err))
	}
}
