package payout

import (
	"context"
	"fmt"
	"time"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	"github.com/google/uuid"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *cronHandler) ReportAfterPayoutCutOffTime(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/payout/ReportAfterPayoutCutOffTime")
	defer segment.End()

	err := error(nil)
	report := disbursementModel.AfterPayoutCutOffTimeSummary{}
	ctx = context.WithValue(ctx, pdkConstant.CtxTraceIdKey, uuid.NewString())

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic Recovery: %v", r)
		}

		status := "FAILED"
		if err == nil {
			status = "SUCCESS"
		}
		h.log.Info(ctx, "Report after payout cut-off time", logger.String("status", status), logger.Any("details", report), logger.Error(err))
	}()

	endTime := time.Now().UTC()
	startTime := endTime.Add(-24 * time.Hour)
	report, err = h.service.ReportAfterPayoutCutOffTime(ctx, startTime, endTime)
}
