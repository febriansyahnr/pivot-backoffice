package merchantHandler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) MigrateBalanceHistoryToDataReporting(ctx context.Context, startDateStr, endDateStr string) error {
	ctx, span := otelTracer.Start(ctx, "port/console/handler/merchant/MigrateBalanceHistoryToDataReporting")
	defer span.End()

	h.logger.Info(ctx, "Running balance history migration to data reporting", logger.String("startDate", startDateStr), logger.String("endDate", endDateStr))

	startDate, err := time.ParseInLocation(time.DateTime, startDateStr+" 00:00:00", tz)
	if err != nil {
		return fmt.Errorf("parse start date: %w", err)
	}

	endDate, err := time.ParseInLocation(time.DateTime, endDateStr+" 23:59:59", tz)
	if err != nil {
		return fmt.Errorf("parse end date: %w", err)
	}

	if startDate.After(endDate) {
		return errors.New("start date must not be after end date")
	}
	return h.reportingSvc.MigrateBalanceHistoryToDataReporting(ctx, startDate.UTC(), endDate.UTC())
}
