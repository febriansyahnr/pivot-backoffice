package payout

import (
	"context"
	"time"

	"github.com/google/uuid"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

// InquirePendingPayout retries the inquiry of pending payout transactions within a specified time range.
// It logs the start and end times of the retry interval, and handles the response from the retry inquiry service.
// If there are any failed transactions, it logs a warning with the total amount and count of failed transactions.
// If the inquiry fails, it logs a fatal error and returns.
func (c *cronHandler) InquirePendingPayout(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/payout/InquirePendingPayout")
	defer segment.End()

	var (
		err error
	)

	ctx = context.WithValue(ctx, pdkConstant.CtxTraceIdKey, uuid.NewString())
	cfg := c.config.DisbursementConfig.RetryInquiryConfig

	endTime := time.Now().UTC().Add(-time.Duration(cfg.DelayTimeMinute) * time.Minute)
	startTime := endTime.Add(-time.Duration(cfg.RetryIntervalMinute) * time.Minute)

	c.log.Info(ctx, "Retry to inquire pending transactions",
		pdkLogger.String("start_time", startTime.String()),
		pdkLogger.String("end_time", endTime.String()),
	)

	summary, err := c.service.RetryInquirePendingTransactions(ctx, startTime, endTime)
	if err != nil {
		c.log.Fatal(ctx, "Failed to inquire pending transactions", pdkLogger.Error(err))
		return
	}

	if summary.TotalFailed > 0 {
		// need to trigger alert
		c.log.Warn(ctx, "Some transactions failed to retry",
			pdkLogger.Float64("total_amount", summary.AmountFailed),
			pdkLogger.Int64("total_failed", summary.TotalFailed),
		)
	}

	c.log.Info(ctx, "Inquiry pending transactions completed", pdkLogger.Any("summary", summary))
}
