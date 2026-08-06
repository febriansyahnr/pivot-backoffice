package partition

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"

	"go.uber.org/zap"
)

func (h *tablePartitionHandler) ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/partition/ReorganizeMonthlyRangePartition")
	defer segment.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		h.logger.Info(
			ctx, "Running reorganize table partition",
			zap.String("tableName", request.TableName),
			zap.String("method", "RANGE"),
			zap.String("durationHuman", duration.String()),
			zap.Int64("durationMs", duration.Milliseconds()),
			zap.Bool("completed", err == nil), zap.Any("error", err),
		)
	}()

	if !strings.EqualFold(request.Datetime.Location().String(), constant.TimeLoc) {
		return fmt.Errorf("calculations and processes using the %s time zone. your timezone %s", constant.TimeLoc, request.Datetime.Location().String())

	} else if request.Datetime.Day() != request.Datetime.EndOfMonth().Day() {
		return fmt.Errorf("operation allowed only at the end of the month. your datetime %s", request.Datetime.Format(time.DateTime))
	}
	return h.partitionService.ReorganizeMonthlyRangePartition(ctx, request)
}
