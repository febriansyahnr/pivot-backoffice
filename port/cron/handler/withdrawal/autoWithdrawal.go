package withdrawalHandler

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *handler) TriggeringAutoWithdrawalProcess(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/withdrawal/TriggeringAutoWithdrawalProcess")
	defer segment.End()

	start := time.Now().UTC()
	messages, err := int64(0), error(nil)

	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	h.logger.Info(ctx, "Running the CronJob to trigger the automatic withdrawal process")

	defer func() {
		duration := time.Now().UTC().Sub(start)

		details := map[string]interface{}{
			"status":        "SUCCESS",
			"duration":      duration.Milliseconds(),
			"durationHuman": duration.String(),
		}
		if err != nil {
			details["status"], details["message"] = "FAILED", err.Error()
		}
		h.logger.Info(ctx, "CronJob completed successfully", zap.Int64("totalMessagesPublished", messages), zap.Any("details", details))
	}()

	if messages, err = h.service.TriggeringAutoWithdrawalProcess(ctx); err != nil {
		h.logger.Fatal(ctx, "Failed when triggering auto withdrawal process", zap.Error(err))
	}
}

// Note: dateStr only for test cases using WIB time
func (h *handler) ForceAutoWithdrawalProcess(ctx context.Context, dateStr string) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/withdrawal/ForceAutoWithdrawalProcess")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	var (
		err    = error(nil)
		start  = time.Now().UTC()
		result = &merchant.ForceAutoWithdrawalProcessResponse{}
	)

	h.logger.Info(ctx, "Starting forced withdrawal process for dormant merchants")
	defer func() {
		duration := time.Now().UTC().Sub(start)

		h.logger.Info(
			ctx, "Forced withdrawal process for dormant merchants completed",
			zap.Any("details", result), zap.Int64("duration", duration.Milliseconds()), zap.String("durationHuman", duration.String()), zap.Bool("completed", err == nil),
		)
	}()

	var date time.Time
	if dateStr == "" {
		date = time.Now().In(tz)

	} else {
		if date, err = time.ParseInLocation(time.DateTime, dateStr, tz); err != nil {
			h.logger.Error(ctx, "Invalid date format. Value: "+dateStr+" must be in YYYY-mm-dd HH:mm:ss format", zap.Error(err))
			return
		}
	}
	date = date.In(time.UTC)

	if result, err = h.service.ForceAutoWithdrawalProcess(ctx, date); err != nil {
		h.logger.Fatal(ctx, "Failed when force auto withdrawal process", zap.Error(err))
	}
}
