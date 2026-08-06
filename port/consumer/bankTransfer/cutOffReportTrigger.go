package bankTransferConsumer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	bankTransferProto "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/bankTransfer"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

// CutOffReportTrigger consumes delayed partner-window triggers and produces one
// report per (partner_code, cutoff window).
//
// Given many delayed messages can arrive for the same window:
// - first, deduplicate each member by external_id so retries/duplicates are ignored,
// - then, index accepted members under one Redis hash for that window,
// - finally, elect a single executor to read indexed members and run report generation.
func (h *handler) CutOffReportTrigger(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/bankTransfer/CutOffReportTrigger")
	defer segment.End()

	payload := &bankTransferProto.CutOffReportTrigger{}
	if err := proto.Unmarshal(body, payload); err != nil {
		return fmt.Errorf("unmarshal cutoff report trigger payload: %w", err)
	}

	if payload.GetPartnerCode() == "" || payload.GetExternalID() == "" || payload.GetCutoffWindowStartAt() == "" || payload.GetCutoffWindowEndAt() == "" {
		return fmt.Errorf("invalid cutoff report trigger payload")
	}

	windowStartAt, err := time.Parse(time.RFC3339, payload.GetCutoffWindowStartAt())
	if err != nil {
		return fmt.Errorf("parse cutoff_window_start_at: %w", err)
	}

	windowEndAt, err := time.Parse(time.RFC3339, payload.GetCutoffWindowEndAt())
	if err != nil {
		return fmt.Errorf("parse cutoff_window_end_at: %w", err)
	}

	windowHash := disbursementService.BuildCutOffReportWindowHash(payload.GetPartnerCode(), windowStartAt, windowEndAt)
	memberDedupKey := fmt.Sprintf(constant.CutOffReportMemberDedupKeyFmt, payload.GetPartnerCode(), windowHash, payload.GetExternalID())
	memberIndexKey := fmt.Sprintf(constant.CutOffReportMemberIndexKeyFmt, payload.GetPartnerCode(), windowHash)
	lastSeenKey := fmt.Sprintf(constant.CutOffReportMemberLastSeenKeyFmt, payload.GetPartnerCode(), windowHash)
	reportDedupKey := fmt.Sprintf(constant.CutOffReportExecutionDedupKeyFmt, payload.GetPartnerCode(), windowHash)
	ttl := time.Duration(constant.CutOffReportRedisTTL) * time.Second
	h.logger.Info(ctx, "[CutOffReportTrigger] trigger received", logger.String("partnerCode", payload.GetPartnerCode()), logger.String("externalID", payload.GetExternalID()), logger.String("windowHash", windowHash))

	isNewMember, err := h.redisExt.SetNX(ctx, memberDedupKey, true, ttl).Result()
	if err != nil {
		return fmt.Errorf("set member dedup key: %w", err)
	}
	// Given at-least-once delivery, duplicate messages are acknowledged as no-op.
	if !isNewMember {
		h.logger.Info(ctx, "[CutOffReportTrigger] duplicate member ignored", logger.String("partnerCode", payload.GetPartnerCode()), logger.String("externalID", payload.GetExternalID()), logger.String("windowHash", windowHash))
		return nil
	}

	if _, err := h.redisExt.HSet(ctx, memberIndexKey, payload.GetExternalID(), 1).Result(); err != nil {
		return fmt.Errorf("add member to index: %w", err)
	}
	h.redisExt.Expire(ctx, memberIndexKey, ttl)

	if err := h.redisExt.Set(ctx, lastSeenKey, time.Now().UTC().UnixNano(), ttl).Err(); err != nil {
		return fmt.Errorf("set last seen key: %w", err)
	}
	h.logger.Info(ctx, "[CutOffReportTrigger] member indexed", logger.String("partnerCode", payload.GetPartnerCode()), logger.String("externalID", payload.GetExternalID()), logger.String("windowHash", windowHash))

	isExecutor, err := h.redisExt.SetNX(ctx, reportDedupKey, true, ttl).Result()
	if err != nil {
		return fmt.Errorf("set report dedup key: %w", err)
	}
	// Then only one execution path becomes report executor for this window.
	if !isExecutor {
		h.logger.Info(ctx, "[CutOffReportTrigger] executor already elected", logger.String("partnerCode", payload.GetPartnerCode()), logger.String("windowHash", windowHash))
		return nil
	}
	h.logger.Info(ctx, "[CutOffReportTrigger] executor elected", logger.String("partnerCode", payload.GetPartnerCode()), logger.String("windowHash", windowHash))

	executorCtx := context.WithoutCancel(ctx)
	go h.generatePartnerWindowReport(executorCtx, payload.GetPartnerCode(), payload.GetPartnerName(), windowHash, windowStartAt.UTC(), windowEndAt.UTC(), memberIndexKey, lastSeenKey)

	return nil
}

func (h *handler) generatePartnerWindowReport(ctx context.Context, partnerCode, partnerName, windowHash string, windowStartAt, windowEndAt time.Time, memberIndexKey, lastSeenKey string) {

	quietWindow := time.Duration(constant.CutOffReportSettleQuietWindowSeconds) * time.Second
	maxWait := time.Duration(constant.CutOffReportSettleMaxWaitSeconds) * time.Second
	deadline := time.Now().Add(maxWait)
	for {
		if time.Now().After(deadline) {
			h.logger.Warn(ctx, "[CutOffReportTrigger] settle wait reached max timeout", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Int64("maxWaitMs", maxWait.Milliseconds()))
			break
		}

		lastSeenRaw, getErr := h.redisExt.Get(ctx, lastSeenKey).Result()
		if getErr != nil {
			h.logger.Warn(ctx, "[CutOffReportTrigger] settle wait skipped due to last seen read error", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Error(getErr))
			break
		}

		lastSeenUnixNano, parseErr := strconv.ParseInt(lastSeenRaw, 10, 64)
		if parseErr != nil {
			h.logger.Warn(ctx, "[CutOffReportTrigger] settle wait skipped due to invalid last seen value", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.String("lastSeenRaw", lastSeenRaw), logger.Error(parseErr))
			break
		}

		age := time.Since(time.Unix(0, lastSeenUnixNano))
		if age >= quietWindow {
			h.logger.Info(ctx, "[CutOffReportTrigger] settle quiet window reached", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Int64("ageMs", age.Milliseconds()))
			break
		}

		waitDuration := quietWindow - age
		remaining := time.Until(deadline)
		if waitDuration > remaining {
			waitDuration = remaining
		}

		if waitDuration <= 0 {
			break
		}

		h.logger.Info(ctx, "[CutOffReportTrigger] waiting for additional members", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Int64("waitMs", waitDuration.Milliseconds()), logger.Int64("ageMs", age.Milliseconds()))

		timer := time.NewTimer(waitDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}

	members := map[string]string{}
	if err := h.redisExt.HGetAllScan(ctx, memberIndexKey, &members); err != nil {
		h.logger.Error(ctx, "[CutOffReportTrigger] read member index failed", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Error(err))
		return
	}
	h.logger.Info(ctx, "[CutOffReportTrigger] final member snapshot", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Int64("memberCount", int64(len(members))))

	externalIDs := make([]string, 0, len(members))
	for externalID := range members {
		externalIDs = append(externalIDs, externalID)
	}

	report, err := h.disbursementSvc.ReportAfterPayoutCutOffTimeByPartnerWindow(ctx, &disbursementModel.PartnerWindowCutOffReportRequest{
		PartnerCode:   partnerCode,
		PartnerName:   partnerName,
		WindowStartAt: windowStartAt.UTC(),
		WindowEndAt:   windowEndAt.UTC(),
		ExternalIDs:   externalIDs,
	})
	if err != nil {
		h.logger.Error(ctx, "[CutOffReportTrigger] partner window report failed", logger.Error(err), logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash))
		return
	}

	h.logger.Info(ctx, "[CutOffReportTrigger] partner window report generated", logger.String("partnerCode", partnerCode), logger.String("windowHash", windowHash), logger.Int64("totalTransaction", report.Total), logger.Any("totalAmount", report.Amount))
}
