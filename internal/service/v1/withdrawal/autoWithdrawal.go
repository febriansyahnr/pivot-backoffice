package withdrawalService

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	commServicePb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"

	ants "github.com/panjf2000/ants/v2"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *withdrawalService) TriggeringAutoWithdrawalProcess(ctx context.Context) (messages int64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/TriggeringAutoWithdrawalProcess")
	defer segment.End()

	merchants, err := s.merchantRepo.GetListOfMerchantsWithActiveAutoWithdrawalStatus(ctx)
	if err != nil {
		return

	} else if len(merchants) == 0 {
		return
	}

	var (
		chanErr = make(chan error, 1)
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workers, _ := ants.NewPoolWithFunc(s.config.AutoWithdrawalWorker, func(merchant interface{}) {
		request, _ := merchant.(merchantModel.MerchantWithActiveAutoWithdrawalStatus)

		intErr := error(nil)

		duration, intErr := s.getTimeDurationPublishAutoWithdrawal(request)

		defer func() {
			chanErr <- intErr

			message := ""
			if intErr != nil {
				message = intErr.Error()
			}
			s.logger.Info(
				ctx, "Preparation for the automatic withdrawal process",
				logger.Any("details", map[string]interface{}{
					"status":       intErr == nil,
					"merchantId":   request.MerchantId,
					"merchantName": request.MerchantName,
					"accountName":  request.AccountName,
					"executeAfter": duration.String(),
					"message":      message,
					"bankCode":     request.BeneficiaryBankCode,
				}),
			)
		}()

		// check available balance
		availableBalance, intErr := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantId, request.AccountName)
		if intErr != nil {
			if intErr != context.Canceled {
				s.logger.Error(ctx, "Get available merchant balance", logger.Error(intErr))
			}
			return

		} else if availableBalance < s.config.MinAmount {
			return
		}

		autoWithdrawalRequest := &pb.WithdrawalRequest{
			MerchantId:           request.MerchantId,
			AccountName:          request.AccountName,
			BeneficiaryBankCode:  request.BeneficiaryBankCode,
			BeneficiaryAccountNo: request.BeneficiaryAccountNo,
			Type:                 constant.WithdrawalAutomated,
			Reason:               constant.WithdrawalReasonScheduled,
			IsFullAmount:         true,
		}
		payload, _ := proto.Marshal(autoWithdrawalRequest)

		if intErr = s.rmq.PublishWithDelay(ctx, rabbitMqExt.WithdrawalProcessRoutingKey, payload, duration); intErr != nil {
			s.logger.Error(ctx, "Failed when publishing message to RabbitMQ", logger.Error(intErr))
			return
		}

		atomic.AddInt64(&messages, 1)
	})
	defer workers.Release()

	go func() {
		for _, merchant := range merchants {
			if ctx.Err() != nil {
				return
			}

			if err := workers.Invoke(merchant); err != nil {
				// Log the error but continue processing other merchants
				continue
			}
		}
	}()

	for range len(merchants) {
		select {
		case <-ctx.Done():
			return messages, ctx.Err()

		case err := <-chanErr:
			if err != nil {
				return messages, err
			}
		}
	}
	return
}

func (s *withdrawalService) ForceAutoWithdrawalProcess(ctx context.Context, date time.Time) (*merchantModel.ForceAutoWithdrawalProcessResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/ForceAutoWithdrawalProcess")
	defer segment.End()

	merchants, err := s.merchantRepo.GetListOfMerchantsToForceTheAutoWithdrawalProcess(ctx)
	if err != nil {
		return nil, err
	}

	wg := new(sync.WaitGroup)
	result := &merchantModel.ForceAutoWithdrawalProcessResponse{Total: len(merchants)}

	workers, _ := ants.NewPoolWithFunc(s.config.AutoWithdrawalWorker, s.forceAutoWithdrawalProcessWorkerFunc(ctx, wg, date, result))
	defer workers.Release()

	for i := range merchants {
		wg.Add(1)
		_ = workers.Invoke(merchants[i])
	}

	wg.Wait()
	return result, nil
}

func (s *withdrawalService) forceAutoWithdrawalProcessWhitelist(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/whitelistForceAutoWithdrawalProcess")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	result, _ := ffclient.BoolVariation(constant.FeatureFlagKeyForceAutoWithdrawalProcess, attr, false)
	return result
}

func (s *withdrawalService) forceAutoWithdrawalProcessWorkerFunc(ctx context.Context, wg *sync.WaitGroup, date time.Time, result *merchantModel.ForceAutoWithdrawalProcessResponse) func(data interface{}) {
	return func(data interface{}) {
		defer wg.Done()

		req := data.(merchantModel.MerchantWithdrawalDetails)

		var (
			err         error
			details     = map[string]string{}
			lastTrxDate = &time.Time{}
			newCtx      = context.WithValue(
				ctx, constant.CtxMerchantIDKey, req.MerchantId,
			)
			availableBalance float64
		)
		defer func() {
			if lastTrxDate == nil {
				lastTrxDate = &time.Time{}
			}
			s.logger.Info(
				newCtx, fmt.Sprintf("Processed Merchant %s - %s (%s)", req.MerchantId, req.MerchantName, req.AccountName),
				logger.Any("details", details), logger.Float64("availableBalance", availableBalance), logger.String("lastTrxDate", lastTrxDate.Format(time.DateTime)+" UTC"),
			)
		}()

		if allowed := s.forceAutoWithdrawalProcessWhitelist(newCtx, req.MerchantId); !allowed {
			atomic.AddInt64(&result.Skip, 1)
			details["message"] = "Non-whitelisted merchant"
			return
		}

		availableBalance, err = s.orchestratorSvc.GetAvailableMerchantBalance(newCtx, req.MerchantId, req.AccountName)
		if err != nil {
			atomic.AddInt64(&result.Failed, 1)
			details["error"], details["message"] = "Get available merchant balances", err.Error()
			return

		} else if availableBalance < s.config.MinAmount {
			atomic.AddInt64(&result.Skip, 1)
			details["message"] = "Merchant balance is less than min transaction"
			return
		}

		lastTrxDate, err = s.accountTrxRepo.GetLastTransactionByAccountName(newCtx, req.MerchantId, req.AccountName)
		if err != nil {
			atomic.AddInt64(&result.Failed, 1)
			details["error"], details["message"] = "Get last transaction by account name", err.Error()
			return
		}
		days := int64(date.Sub(*lastTrxDate).Hours() / 24)

		withdrawalAfterInactivityDays := s.config.PaymentBalanceConfig.WithdrawalAfterInactivityDays
		notificationAfterInactivityDays := s.config.PaymentBalanceConfig.NotificationAfterInactivityDays
		switch req.AccountName {
		case constant.AccountNameDisbursement:
			notificationAfterInactivityDays = s.config.DisbursementBalanceConfig.NotificationAfterInactivityDays
			withdrawalAfterInactivityDays = s.config.DisbursementBalanceConfig.WithdrawalAfterInactivityDays

		case constant.AccountNameVirtualTerminal:
			notificationAfterInactivityDays = s.config.VirtualTerminalBalanceConfig.NotificationAfterInactivityDays
			withdrawalAfterInactivityDays = s.config.VirtualTerminalBalanceConfig.WithdrawalAfterInactivityDays
		}

		if days == int64(notificationAfterInactivityDays) {
			atomic.AddInt64(&result.Notify, 1)
			details["message"] = "Send notification email"

			content, _ := structpb.NewStruct(map[string]any{
				"LogoURL":                config.GetEmailLogoURL(),
				"MerchantName":           req.MerchantName,
				"AccountName":            strings.ToLower(req.AccountName),
				"Days":                   days,
				"BeneficiaryBankName":    req.BeneficiaryBankName,
				"BeneficiaryAccountNo":   req.BeneficiaryAccountNo,
				"BeneficiaryAccountName": req.BeneficiaryAccountName,
			})
			emailRequest := &commServicePb.EmailRequest{
				Event:    constant.ForceAutoWithdrawalNotifyEvent,
				From:     config.DefaultEmailSender(),
				To:       req.MerchantEmail,
				Subject:  "[IMPORTANT] Active Balance Automatic Refund to Registered Bank Account Reminder",
				Content:  content,
				Priority: commServicePb.EmailPriority_L0, ToBeRetriedOnFailure: true,
			}
			payload, _ := proto.Marshal(emailRequest)

			_ = s.rmq.Publish(newCtx, rabbitMqExt.CommServiceEmailRoutingKey, nil, payload)

			s.logger.Info(ctx, "Send notification email to merchant", logger.String("merchantId", req.MerchantId), logger.String("merchantName", req.MerchantName), logger.String("accountName", req.AccountName), logger.Int64("inactiveDays", days))

		} else if days >= int64(withdrawalAfterInactivityDays) {
			atomic.AddInt64(&result.Dormant, 1)
			details["message"] = "Publish auto withdrawal process"

			autoWithdrawalRequest := &pb.WithdrawalRequest{
				MerchantId:           req.MerchantId,
				AccountName:          req.AccountName,
				BeneficiaryBankCode:  req.BeneficiaryBankCode,
				BeneficiaryAccountNo: req.BeneficiaryAccountNo,
				Amount:               availableBalance,
				Type:                 constant.WithdrawalAutomated,
				Reason:               constant.WithdrawalReasonDormantMerchant,
			}
			payload, _ := proto.Marshal(autoWithdrawalRequest)

			_ = s.rmq.Publish(newCtx, rabbitMqExt.WithdrawalProcessRoutingKey, nil, payload)

			s.logger.Info(ctx, "Balance merchant withdraw", logger.String("merchantId", req.MerchantId), logger.String("merchantName", req.MerchantName), logger.String("accountName", req.AccountName), logger.Int64("inactiveDays", days))

		} else {
			atomic.AddInt64(&result.Skip, 1)
			details["message"] = fmt.Sprintf("Skipped, last transaction %d days ago", days)
		}
	}
}

func (s *withdrawalService) getTimeDurationPublishAutoWithdrawal(request merchantModel.MerchantWithActiveAutoWithdrawalStatus) (time.Duration, error) {

	if slices.Contains(s.config.AutoWithdrawalBankCodeToBeExecAfterTrigger, request.BeneficiaryBankCode) {
		// If bank code is in the trigger list, execute immediately
		return 0, nil
	}

	// parse time from config
	now, err := util.GetJakartaTime()
	if err != nil {
		return 0, err
	}

	parsedTime, err := time.Parse("15:04:05", s.config.AutoWithdrawalDefaultSchedulingTime)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(0)
	// if current time is before the scheduling time, set duration to the scheduling time today
	if now.Hour() < parsedTime.Hour() ||
		(now.Hour() == parsedTime.Hour() && now.Minute() < parsedTime.Minute()) ||
		(now.Hour() == parsedTime.Hour() && now.Minute() == parsedTime.Minute() && now.Second() < parsedTime.Second()) {
		duration = time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, now.Location()).Sub(now)

	} else {
		// if current time is after the scheduling time, set duration to the scheduling time tomorrow
		duration = time.Date(now.Year(), now.Month(), now.Day()+1, parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, now.Location()).Sub(now)
	}

	return duration, nil
}
