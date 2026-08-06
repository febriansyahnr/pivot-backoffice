package payment

import (
	"context"
	"time"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"go.uber.org/zap"
)

func (h *paymentCronHandler) ProcessInvestigationMonthlyReconciliation(ctx context.Context, dateStr string) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/payment/ProcessInvestigationMonthlyReconciliation")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

	var (
		err     error
		start   = time.Now()
		request model.MonthlyReconciliationRequest
	)

	h.logger.Info(ctx, "Starting payment investigation monthly reconciliation")

	var date time.Time
	if dateStr == "" {
		date = time.Now().In(localTZ)

	} else {
		if date, err = time.ParseInLocation(time.DateTime, dateStr, localTZ); err != nil {
			h.logger.Error(ctx, "Invalid date format. Value: "+dateStr+" must be in YYYY-mm-dd HH:mm:ss format", zap.Error(err))
			return
		}
	}

	defer func() {
		duration := time.Since(start)

		h.logger.Info(
			ctx, "Payment investigation monthly reconciliation completed",
			zap.Any("request", request), zap.Int64("duration", duration.Milliseconds()), zap.String("durationHuman", duration.String()), zap.Bool("completed", err == nil), zap.Error(err),
		)
	}()

	// The reconciliation process runs every early morning and includes transactions from the same date of the previous month until the end of the previous day.
	// The start and end of day are calculated using the Asia/Jakarta (GMT+7) local timezone and then converted to UTC.
	request = model.MonthlyReconciliationRequest{
		StartDate: time.Date(date.Year(), date.Month()-1, date.Day(), 0, 0, 0, 0, date.Location()).UTC(),
		EndDate:   time.Date(date.Year(), date.Month(), date.Day()-1, 23, 59, 59, 0, date.Location()).UTC(),
	}
	err = h.paymentService.ProcessInvestigationMonthlyReconciliation(ctx, request)
}
