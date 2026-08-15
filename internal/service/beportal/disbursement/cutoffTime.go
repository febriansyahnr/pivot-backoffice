package disbursementService

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"

	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"google.golang.org/protobuf/proto"
)

func (s *DisbursementService) GetCutOffTimeStatus(ctx context.Context, now time.Time, merchantId string, windowConfig *config.DisbursementCutOffTimeWindow) (result *disbursementModel.CutOffTimeStatusResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetCutOffTimeStatus")
	defer segment.End()

	config := s.config.DisbursementConfig.CutOffTimeWindow
	if windowConfig != nil {
		config = *windowConfig
	}

	result = &disbursementModel.CutOffTimeStatusResponse{
		Status: constant.DisbursementCutOffTimeStatusDeactive,
	}
	if !config.Enabled {
		return
	}

	tz := time.FixedZone("GMT", config.GMT*60*60)

	now = now.In(tz)

	startTimeWindow, err := time.ParseInLocation(time.DateTime, now.Format(time.DateOnly)+" "+config.StartTime+":00", tz)
	if err != nil {
		s.logger.Error(ctx, "parse payout cut-off start time window", logger.Error(err))
		return nil, fmt.Errorf("parse start time window: %v", err)
	}

	endTimeWindow, err := time.ParseInLocation(time.DateTime, now.Format(time.DateOnly)+" "+config.EndTime+":59", tz)
	if err != nil {
		s.logger.Error(ctx, "parse payout cut-off end time window", logger.Error(err))
		return nil, fmt.Errorf("parse end time window: %v", err)
	}

	result.Time = now.Format(time.RFC3339)

	// With Cross Day Config
	if !config.SameDay {
		for _, add := range [][]int{{-1, 0}, {0, 1}} {
			t1 := startTimeWindow.AddDate(0, 0, add[0])
			t2 := endTimeWindow.AddDate(0, 0, add[1])

			if t1.Before(now) && t2.After(now) {
				if s.doNotApplyPayoutCutOffTime(ctx, merchantId) {

					result.Status = constant.DisbursementCutOffTimeStatusWhitelisted
					return
				}

				result.Banner = config.BannerStatus
				result.Status = constant.DisbursementCutOffTimeStatusOngoing
				result.ProcessedAt = t2.Add(time.Duration(config.TimeLagForSendingReportSecond) * time.Second).UTC()
				return
			}
		}
		// Off Schedule For Cros Day Config
		result.Status = constant.DisbursementCutOffTimeStatusOffSchedule
		return
	}

	if now.Before(startTimeWindow) || now.After(endTimeWindow) {

		result.Status = constant.DisbursementCutOffTimeStatusOffSchedule
		return

	} else if s.doNotApplyPayoutCutOffTime(ctx, merchantId) {

		result.Status = constant.DisbursementCutOffTimeStatusWhitelisted
		return
	}

	result.Banner = config.BannerStatus
	result.Status = constant.DisbursementCutOffTimeStatusOngoing
	result.ProcessedAt = endTimeWindow.Add(time.Duration(config.TimeLagForSendingReportSecond) * time.Second).UTC()
	return
}

func (s *DisbursementService) doNotApplyPayoutCutOffTime(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/doNotApplyPayoutCutOffTime")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	result, _ := ffclient.BoolVariation(constant.FeatureFlagKeyDoNotApplyPayoutCutOffTime, attr, true)
	return result
}

// This cron is used to send a summary of pending transaction reports (via Slack) during the Payout Cut-off Time.
//
// Start and end times are only used when data in the cache is not found.
func (s *DisbursementService) ReportAfterPayoutCutOffTime(ctx context.Context, startTime, endTime time.Time) (report disbursementModel.AfterPayoutCutOffTimeSummary, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ReportAfterPayoutCutOffTime")
	defer segment.End()

	result := map[string]string{}

	if err = s.redisExt.HGetAllScan(ctx, constant.DelayTransferProcessRedisKey, &result); err != nil && !errors.Is(err, redisExt.ErrNil) {
		return report, fmt.Errorf("Scanning data from cache: %v", err)
	}
	s.logger.Info(ctx, "Data sourced from cache", logger.Any("result", result))

	if len(result) == 0 {
		report, err = s.disbursementRepo.GetSummaryOfDelayedTransactionBeforeProcessed(ctx, startTime, endTime)
		if err != nil {
			return report, fmt.Errorf("Get summary of delayed transaction before processed: %v", err)
		}

	} else {
		s.mappingReportAfterPayoutCutOffTime(result, &report)
	}

	var slaDuration, avgDuration time.Duration
	if report.Total > 0 {
		t2 := time.Now().UTC()
		t1 := t2.AddDate(0, -1, 0)

		avgDurationInMs, err := s.disbursementRepo.GetAvgDurationOfBankTransferProcessInMs(ctx, t1, t2)
		if err != nil {
			return report, fmt.Errorf("Get avg duration of bank transfer in ms: %w", err)
		}
		avgDuration = time.Duration(avgDurationInMs) * time.Millisecond
		slaDuration = time.Duration(report.Total) * avgDuration
	}
	report.Info = fmt.Sprintf("SLA Estimation : %s Avg. Process : %s", slaDuration.String(), avgDuration.String())

	boundaryLine := strings.Repeat("-", 80)
	message := &slackPb.PostWebhookCmd{
		URL:   s.config.SlackConfig.PayoutCutOffTimeWebHookURL,
		Color: slackPb.Color_GOOD,
		Title: "Payout Cut-off Time Report",
	}
	attachmentFields := []*slackPb.AttachmentField{
		{Title: "Transaction Summary"},
		{Value: fmt.Sprintf("Total Transaction : %d", report.Total)},
		{Value: fmt.Sprintf("Total Amount       : Rp. %s", util.ConvertFloatToCurrency(report.Amount))},
		{Value: fmt.Sprintf("SLA Estimation    : %s    Avg. Process : %s", slaDuration.String(), avgDuration.String())},
		{Value: ""},
		{Title: "Bank Details"},
		{Value: boundaryLine},
	}
	if len(report.Banks) == 0 {
		attachmentFields = append(attachmentFields,
			&slackPb.AttachmentField{
				Value: "No transaction details",
			},
			&slackPb.AttachmentField{Value: boundaryLine, Short: false},
		)
	}
	for _, bank := range report.Banks {
		if len(bank.Name) > 30 {
			bank.Name = bank.Name[:30]
		}
		attachmentFields = append(attachmentFields,
			&slackPb.AttachmentField{
				Value: fmt.Sprintf("*Bank Name* : %s", bank.Name), Short: false,
			},
			&slackPb.AttachmentField{
				Value: fmt.Sprintf("*Total*             : %d      *Amount* : Rp. %s", bank.Total, util.ConvertFloatToCurrency(bank.Amount)), Short: false,
			},
			&slackPb.AttachmentField{Value: boundaryLine, Short: false},
		)
	}
	attachmentFields = append(attachmentFields,
		&slackPb.AttachmentField{
			Value: fmt.Sprintf("Generate Date: %s", time.Now().In(local).Format(time.RFC850)),
		},
		&slackPb.AttachmentField{
			Value: boundaryLine, Short: false,
		},
	)
	message.Fields = attachmentFields

	rawPayload, _ := proto.Marshal(message)

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawPayload)
	return
}

func (s *DisbursementService) mappingReportAfterPayoutCutOffTime(result map[string]string, dst *disbursementModel.AfterPayoutCutOffTimeSummary) {
	if result == nil || dst == nil {
		return
	}

	// Example Result:
	// map[total: 3 amount: 30000 bank_002_total: 1 bank_002_amount: 10000 bank_014_total: 2 bank_014_amount: 20000]
	indexes := map[string]int{}
	replacer := strings.NewReplacer("bank_", "", "_total", "", "_amount", "")

	// Parsing Result Value
	dst.Total, _ = strconv.ParseInt(result["total"], 10, 64)
	dst.Amount, _ = strconv.ParseFloat(result["amount"], 64)
	dst.Banks = make([]disbursementModel.AfterPayoutCutOffTimeBankSummary, 0, ((len(result) - 2) / 2))

	for k, v := range result {
		if k == "amount" || k == "total" {
			continue
		}
		bankCode := replacer.Replace(k)

		i, ok := indexes[bankCode]
		if !ok {
			bank := bankDB.FindByCode(bankCode)
			if bank == nil {
				bank = &bankTransfer.Bank{}
			}
			dst.Banks = append(dst.Banks, disbursementModel.AfterPayoutCutOffTimeBankSummary{Name: bank.Name})

			i = len(dst.Banks) - 1
			indexes[bankCode] = i
		}

		val, _ := strconv.ParseFloat(v, 64)
		if strings.HasSuffix(k, "_total") {
			dst.Banks[i].Total += int64(val)

		} else if strings.HasSuffix(k, "_amount") {
			dst.Banks[i].Amount += val
		}
	}
}

// ReportAfterPayoutCutOffTimeByPartnerWindow resolves partner-window external IDs
// into disbursements, aggregates summary fields, and publishes one Slack report.
func (s *DisbursementService) ReportAfterPayoutCutOffTimeByPartnerWindow(ctx context.Context, req *disbursementModel.PartnerWindowCutOffReportRequest) (report disbursementModel.AfterPayoutCutOffTimeSummary, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ReportAfterPayoutCutOffTimeByPartnerWindow")
	defer segment.End()

	if req == nil || req.PartnerCode == "" || len(req.ExternalIDs) == 0 {
		return report, nil
	}

	banks := map[string]*disbursementModel.AfterPayoutCutOffTimeBankSummary{}
	merchants := map[string]struct{}{}
	disbursementSeen := map[string]struct{}{}

	for _, externalID := range req.ExternalIDs {
		transaction, trxErr := s.accountTransactionRepo.FindByID(ctx, externalID)
		if trxErr != nil || transaction == nil {
			s.logger.Warn(ctx, "[CutOffReport] failed resolve account transaction", logger.String("externalID", externalID), logger.Error(trxErr))
			continue
		}

		if transaction.Type != constant.TypeDisbursement && transaction.Type != constant.TypeBulkDisbursement {
			continue
		}

		disbursementID := transaction.ReferenceID
		if disbursementID == "" {
			continue
		}
		if _, ok := disbursementSeen[disbursementID]; ok {
			continue
		}

		disbursement, disbErr := s.disbursementRepo.FindByID(ctx, disbursementID)
		if disbErr != nil || disbursement == nil {
			s.logger.Warn(ctx, "[CutOffReport] failed resolve disbursement", logger.String("disbursementID", disbursementID), logger.Error(disbErr))
			continue
		}

		disbursementSeen[disbursementID] = struct{}{}
		report.Total++
		report.Amount += disbursement.Amount.InexactFloat64()
		merchants[disbursement.MerchantID] = struct{}{}

		bankName := "UNKNOWN"
		if disbursement.BeneficiaryBankName != nil && *disbursement.BeneficiaryBankName != "" {
			bankName = *disbursement.BeneficiaryBankName
		}

		if banks[bankName] == nil {
			banks[bankName] = &disbursementModel.AfterPayoutCutOffTimeBankSummary{Name: bankName}
		}
		banks[bankName].Total++
		banks[bankName].Amount += disbursement.Amount.InexactFloat64()
	}

	report.Banks = make([]disbursementModel.AfterPayoutCutOffTimeBankSummary, 0, len(banks))
	for _, bank := range banks {
		report.Banks = append(report.Banks, *bank)
	}
	sort.Slice(report.Banks, func(i, j int) bool {
		return report.Banks[i].Name < report.Banks[j].Name
	})

	report.Merchants = make([]string, 0, len(merchants))
	for merchantID := range merchants {
		report.Merchants = append(report.Merchants, merchantID)
	}
	sort.Strings(report.Merchants)

	var slaDuration, avgDuration time.Duration
	if report.Total > 0 {
		t2 := time.Now().UTC()
		t1 := t2.AddDate(0, -1, 0)

		avgDurationInMs, avgErr := s.disbursementRepo.GetAvgDurationOfBankTransferProcessInMs(ctx, t1, t2)
		if avgErr != nil {
			return report, fmt.Errorf("Get avg duration of bank transfer in ms: %w", avgErr)
		}
		avgDuration = time.Duration(avgDurationInMs) * time.Millisecond
		slaDuration = time.Duration(report.Total) * avgDuration
	}
	report.Info = fmt.Sprintf("SLA Estimation : %s Avg. Process : %s", slaDuration.String(), avgDuration.String())

	title := fmt.Sprintf("Payout Cut-off Time Report [%s - %s]", req.PartnerName, req.PartnerCode)
	boundaryLine := strings.Repeat("-", 80)
	message := &slackPb.PostWebhookCmd{
		URL:   s.config.SlackConfig.PayoutCutOffTimeWebHookURL,
		Color: slackPb.Color_GOOD,
		Title: title,
	}

	windowLabel := fmt.Sprintf("%s - %s", req.WindowStartAt.In(local).Format(time.RFC3339), req.WindowEndAt.In(local).Format(time.RFC3339))
	attachmentFields := []*slackPb.AttachmentField{
		{Title: "Partner Window"},
		{Value: fmt.Sprintf("Partner Code      : %s", req.PartnerCode)},
		{Value: fmt.Sprintf("Partner Name      : %s", req.PartnerName)},
		{Value: fmt.Sprintf("Window            : %s", windowLabel)},
		{Value: ""},
		{Title: "Transaction Summary"},
		{Value: fmt.Sprintf("Total Transaction : %d", report.Total)},
		{Value: fmt.Sprintf("Total Amount       : Rp. %s", util.ConvertFloatToCurrency(report.Amount))},
		{Value: fmt.Sprintf("SLA Estimation    : %s    Avg. Process : %s", slaDuration.String(), avgDuration.String())},
		{Value: ""},
		{Title: "Merchant UUIDs"},
	}

	if len(report.Merchants) == 0 {
		attachmentFields = append(attachmentFields, &slackPb.AttachmentField{Value: "No merchant data"})
	} else {
		for _, merchantID := range report.Merchants {
			attachmentFields = append(attachmentFields, &slackPb.AttachmentField{Value: merchantID, Short: false})
		}
	}

	attachmentFields = append(attachmentFields,
		&slackPb.AttachmentField{Value: ""},
		&slackPb.AttachmentField{Title: "Bank Details"},
		&slackPb.AttachmentField{Value: boundaryLine},
	)

	if len(report.Banks) == 0 {
		attachmentFields = append(attachmentFields,
			&slackPb.AttachmentField{Value: "No transaction details"},
			&slackPb.AttachmentField{Value: boundaryLine, Short: false},
		)
	}

	for _, bank := range report.Banks {
		name := bank.Name
		if len(name) > 30 {
			name = name[:30]
		}
		attachmentFields = append(attachmentFields,
			&slackPb.AttachmentField{Value: fmt.Sprintf("*Bank Name* : %s", name), Short: false},
			&slackPb.AttachmentField{Value: fmt.Sprintf("*Total*             : %d      *Amount* : Rp. %s", bank.Total, util.ConvertFloatToCurrency(bank.Amount)), Short: false},
			&slackPb.AttachmentField{Value: boundaryLine, Short: false},
		)
	}

	attachmentFields = append(attachmentFields,
		&slackPb.AttachmentField{Value: fmt.Sprintf("Generate Date: %s", time.Now().In(local).Format(time.RFC850))},
		&slackPb.AttachmentField{Value: boundaryLine, Short: false},
	)
	message.Fields = attachmentFields

	rawPayload, _ := proto.Marshal(message)
	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawPayload)

	return report, nil
}

func BuildCutOffReportWindowHash(partnerCode string, startAt, endAt time.Time) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s", partnerCode, startAt.UTC().Format(time.RFC3339), endAt.UTC().Format(time.RFC3339))))
	return hex.EncodeToString(hash[:])
}
