package partition

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/tablePartitionExt"
	"go.uber.org/zap"
)

// CreateAccountTransactionPartition return nothing, it will create partition for account transactions table
// it will create 15 partition for 15 days ahead at 18:00 UTC
func (h *tablePartitionHandler) CreateAccountTransactionPartition(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/CreateAccountTransactionPartition")
	defer segment.End()

	now := time.Now().UTC()

	err := h.tablePartition.CreateDayRangePartition(ctx, tablePartitionExt.PartitionConfig{
		TableName:          "account_transactions",
		TotalPartition:     15,
		StartedAt:          time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC),
		Parameter:          "created_at",
		IsPreciseTimestamp: true,
	})

	if err != nil {
		h.logger.Fatal(ctx, "err-create-partition--account_transactions:", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "account_transactions partition creation succeeded")
}

// CreateCallbackLogPartition return nothing, it will create partition for callback_logs table
// it will create 15 partition for 15 days ahead at 18:00 UTC
func (h *tablePartitionHandler) CreateCallbackLogPartition(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/CreateCallbackLogPartition")
	defer segment.End()

	now := time.Now().UTC()

	err := h.tablePartition.CreateDayRangePartition(ctx, tablePartitionExt.PartitionConfig{
		TableName:      "callback_logs",
		TotalPartition: 15,
		StartedAt:      time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC),
		Parameter:      "created_at",
	})

	if err != nil {
		h.logger.Fatal(ctx, "err-create-partition--callback_logs:", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "callback_logs partition creation succeeded")
}

// CreatePaymentPartition return nothing, it will create partition for payments table
// it will create 15 partition for 15 days ahead at 18:00 UTC
func (h *tablePartitionHandler) CreatePaymentPartition(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/CreatePaymentPartition")
	defer segment.End()

	now := time.Now().UTC()

	err := h.tablePartition.CreateDayRangePartition(ctx, tablePartitionExt.PartitionConfig{
		TableName:      "payments",
		TotalPartition: 15,
		StartedAt:      time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC),
		Parameter:      "created_at",
	})

	if err != nil {
		h.logger.Fatal(ctx, "err-create-partition--payments:", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "payments partition creation succeeded")
}

// CreateDisbursementPartition return nothing, it will create partition for disbursement table
// it will create 15 partition for 15 days ahead at 18:00 UTC
func (h *tablePartitionHandler) CreateDisbursementPartition(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/CreateDisbursementPartition")
	defer segment.End()

	now := time.Now().UTC()

	err := h.tablePartition.CreateDayRangePartition(ctx, tablePartitionExt.PartitionConfig{
		TableName:      "disbursements",
		TotalPartition: 15,
		StartedAt:      time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC),
		Parameter:      "created_at",
	})

	if err != nil {
		h.logger.Fatal(ctx, "err-create-partition--disbursements:", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "disbursements partition creation succeeded")
}

// CreateActivityLogPartition return nothing, it will create partition for activity_logs table
// it will create 15 partition for 15 days ahead at 18:00 UTC
func (h *tablePartitionHandler) CreateActivityLogPartition(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/CreateActivityLogPartition")
	defer segment.End()

	now := time.Now().UTC()

	err := h.tablePartition.CreateDayRangePartition(ctx, tablePartitionExt.PartitionConfig{
		TableName:      "activity_logs",
		TotalPartition: 15,
		StartedAt:      time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC),
		Parameter:      "created_at",
	})

	if err != nil {
		h.logger.Fatal(ctx, "err-create-partition--activity_logs:", zap.Error(err))
		return
	}

	h.logger.Info(ctx, "activity_logs partition creation succeeded")
}
