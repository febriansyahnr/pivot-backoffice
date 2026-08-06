package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

// DecrBeneficiaryPayoutLimit is a function to decrement the beneficiary payout limit for default, custom, and merchant policy rules
func (s *DisbursementService) DecrBeneficiaryPayoutLimit(ctx context.Context, merchantID, bankCode, accountNo string, amount float64) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/DecrDailyTransactionLimit")
	defer segment.End()

	if derivedMerchantID, _ := ctx.Value(constant.CtxDerivedMerchantID).(string); derivedMerchantID != "" {
		merchantID = derivedMerchantID
	}

	customCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo,
	)
	merchantPolicyCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantID, bankCode, accountNo,
	)
	defaultCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo,
	)

	// Check if custom cache key exists
	res := s.redisExt.Exists(ctx, customCacheKey)
	customCacheKeyExist := res.Val() > 0
	err := res.Err()
	if err != nil {
		s.logger.Error(ctx, "Fail on checking custom cache key existence", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	// If custom rule exists, use it
	if customCacheKeyExist {
		if _, err := s.redisExt.HIncrByFloat(ctx, customCacheKey, "processed", -amount); err != nil {
			s.logger.Error(ctx, "Decrement processed amount for custom rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		if _, err := s.redisExt.HIncrBy(ctx, customCacheKey, "count", -1); err != nil {
			s.logger.Error(ctx, "Decrement count for custom rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		s.logger.Info(ctx, "[DecrBeneficiaryPayoutLimit] Restore beneficiary payout custom rule limit", logger.Any("details", map[string]any{
			"merchantID": merchantID,
			"bankCode":   bankCode,
			"accountNo":  accountNo,
			"amount":     amount,
		}))
		return nil
	}

	// Check if merchant policy cache key exists
	res = s.redisExt.Exists(ctx, merchantPolicyCacheKey)
	merchantPolicyCacheKeyExist := res.Val() > 0
	err = res.Err()
	if err != nil {
		s.logger.Error(ctx, "Fail on checking merchant policy cache key existence", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	// If merchant policy rule exists, use it
	if merchantPolicyCacheKeyExist {
		if _, err := s.redisExt.HIncrByFloat(ctx, merchantPolicyCacheKey, "processed", -amount); err != nil {
			s.logger.Error(ctx, "Decrement processed amount for merchant policy rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		if _, err := s.redisExt.HIncrBy(ctx, merchantPolicyCacheKey, "count", -1); err != nil {
			s.logger.Error(ctx, "Decrement count for merchant policy rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		s.logger.Info(ctx, "[DecrBeneficiaryPayoutLimit] Restore beneficiary payout merchant policy rule limit", logger.Any("details", map[string]any{
			"merchantID": merchantID,
			"bankCode":   bankCode,
			"accountNo":  accountNo,
			"amount":     amount,
		}))
		return nil
	}

	s.logger.Info(ctx, "Custom and merchant policy rules not found, looking for default rule", logger.Any("details", map[string]any{
		"merchantID":             merchantID,
		"bankCode":               bankCode,
		"accountNo":              accountNo,
		"amount":                 amount,
		"customCacheKey":         customCacheKey,
		"merchantPolicyCacheKey": merchantPolicyCacheKey,
		"defaultCacheKey":        defaultCacheKey,
	}))

	// If custom and merchant policy rules don't exist, try default rule
	res = s.redisExt.Exists(ctx, defaultCacheKey)
	defaultCacheKeyExists := res.Val() > 0
	err = res.Err()
	if err != nil {
		s.logger.Error(ctx, "Fail on checking default cache key existence", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if defaultCacheKeyExists {
		if _, err := s.redisExt.HIncrByFloat(ctx, defaultCacheKey, "processed", -amount); err != nil {
			s.logger.Error(ctx, "Decrement processed amount for default rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		if _, err := s.redisExt.HIncrBy(ctx, defaultCacheKey, "count", -1); err != nil {
			s.logger.Error(ctx, "Decrement count for default rule", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
		s.logger.Info(ctx, "[DecrBeneficiaryPayoutLimit] Restore beneficiary payout default rule limit", logger.Any("details", map[string]any{
			"bankCode":  bankCode,
			"accountNo": accountNo,
			"amount":    amount,
		}))
		return nil
	}

	// None of the rules exist
	s.logger.Warn(ctx, "Failed to find custom, merchant policy, or default rule cache key", logger.Any("details", map[string]any{
		"merchantID":             merchantID,
		"bankCode":               bankCode,
		"accountNo":              accountNo,
		"customCacheKey":         customCacheKey,
		"merchantPolicyCacheKey": merchantPolicyCacheKey,
		"defaultCacheKey":        defaultCacheKey,
	}))

	return nil
}

func (s *DisbursementService) ValidateBeneficiaryPayoutDefaultRule(
	ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateBeneficiaryPayoutCustomRule")
	defer segment.End()

	s.logger.Info(ctx, "validating beneficiary payout default rule",
		logger.String("bankCode", bankCode),
		logger.String("accountNo", accountNo),
		logger.Float64("amount", amount),
	)

	var (
		processedValue = 0.0
		count          = int64(0)
	)
	limitResp := &disbursementModel.BeneficiaryPayoutLimitRuleLimit{}
	if rule != nil {
		limitResp.BeneficiaryPayoutLimitRuleConfig = *rule
	}

	cacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo,
	)

	err := s.redisExt.HGetAllScan(ctx, cacheKey, limitResp)
	if err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Get beneficiary payout default rule limit from cache", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	}

	if errors.Is(err, redisExt.ErrNil) || limitResp.Count == 0 {
		// when cache not found or got 0 counter
		// re-validate by querying the database
		// should provide empty merchant to calculate all payout trx
		payoutLimit, err := s.GetBeneficiaryPayoutRuleLimit(ctx, "", bankCode, accountNo, rule)
		if err != nil {
			s.logger.Error(ctx, "failed to re-validate beneficiary limit", logger.Error(err))
			return err
		}

		processedValue = payoutLimit.Processed
		count = int64(payoutLimit.Count)
	} else {
		processedValue, count, err = s.updateBeneficiaryQuota(ctx, cacheKey, amount)
		if err != nil {
			s.logger.Error(ctx, "failed to reduce beneficiary limit quota", logger.Error(err))
			return err
		}
	}

	isProcessedValueWithinLimit := processedValue > s.config.DisbursementConfig.BeneficiaryLimit.Amount
	isVelocityWithinLimit := count > s.config.DisbursementConfig.BeneficiaryLimit.Velocity
	if isProcessedValueWithinLimit || isVelocityWithinLimit {
		_, _ = s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", -amount)
		_, _ = s.redisExt.HIncrBy(ctx, cacheKey, "count", -1)

		err := pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions)

		s.sendBeneficiaryPayoutLimitAlert(ctx, disbursementModel.BeneficiaryPayoutLimitAlertRequest{
			TotalAmount:              processedValue,
			NumberOfTransaction:      count,
			BeneficiaryBankCode:      bankCode,
			BeneficiaryAccountNumber: accountNo,
			BeneficiaryAccountName:   accountName,
			MerchantID:               merchantId,
			AmountThreshold:          s.config.DisbursementConfig.BeneficiaryLimit.Amount,
			CountThreshold:           s.config.DisbursementConfig.BeneficiaryLimit.Velocity,
		})

		s.logger.Error(ctx, "[ValidateBeneficiaryPayoutDefaultRule] beneficiary limit exceeded", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "[ValidateBeneficiaryPayoutDefaultRule] Reduce beneficiary payout default rule limit", logger.Any("details", map[string]any{
		"bankCode":  bankCode,
		"accountNo": accountNo,
		"amount":    amount,
	}))

	return nil
}

func (s *DisbursementService) ValidateBeneficiaryPayoutCustomRule(
	ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateBeneficiaryPayoutCustomRule")
	defer segment.End()

	s.logger.Info(ctx, "validating beneficiary payout custom rule",
		logger.String("merchantId", merchantId),
		logger.String("bankCode", bankCode),
		logger.String("accountNo", accountNo),
		logger.Float64("amount", amount),
	)

	var (
		processedValue = 0.0
		count          = int64(0)
	)
	limitResp := &disbursementModel.BeneficiaryPayoutLimitRuleLimit{}
	if rule != nil {
		limitResp.BeneficiaryPayoutLimitRuleConfig = *rule
	}

	cacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantId, bankCode, accountNo,
	)
	err := s.redisExt.HGetAllScan(ctx, cacheKey, limitResp)
	if err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Get beneficiary payout custom rule limit from cache", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if errors.Is(err, redisExt.ErrNil) || limitResp.Count == 0 {
		// when cache not found or got 0 counter
		// re-validate by querying the database
		payoutLimit, err := s.GetBeneficiaryPayoutRuleLimit(ctx, merchantId, bankCode, accountNo, rule)
		if err != nil {
			s.logger.Error(ctx, "failed to re-validate beneficiary limit", logger.Error(err))
			return err
		}

		processedValue = payoutLimit.Processed
		count = int64(payoutLimit.Count)
	} else {
		processedValue, count, err = s.updateBeneficiaryQuota(ctx, cacheKey, amount)
		if err != nil {
			s.logger.Error(ctx, "failed to reduce beneficiary limit quota", logger.Error(err))
			return err
		}
	}

	if processedValue > limitResp.AmountThreshold || count > limitResp.Velocity {
		_, _ = s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", -amount)
		_, _ = s.redisExt.HIncrBy(ctx, cacheKey, "count", -1)

		err := pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions)

		s.sendBeneficiaryPayoutLimitAlert(ctx, disbursementModel.BeneficiaryPayoutLimitAlertRequest{
			TotalAmount:              processedValue,
			NumberOfTransaction:      count,
			BeneficiaryBankCode:      bankCode,
			BeneficiaryAccountNumber: accountNo,
			BeneficiaryAccountName:   accountName,
			MerchantID:               merchantId,
			AmountThreshold:          limitResp.AmountThreshold,
			CountThreshold:           limitResp.Velocity,
		})

		s.logger.Error(ctx, "[ValidateBeneficiaryPayoutCustomRule] beneficiary limit exceeded", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "[ValidateBeneficiaryPayoutCustomRule] Reduce beneficiary payout custom rule limit", logger.Any("details", map[string]any{
		"merchantID": merchantId,
		"bankCode":   bankCode,
		"accountNo":  accountNo,
		"rule":       rule,
		"amount":     amount,
	}))

	return nil
}

func (s *DisbursementService) ValidateBeneficiaryPayoutMerchantPolicyRule(
	ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateBeneficiaryPayoutMerchantPolicyRule")
	defer segment.End()

	s.logger.Info(ctx, "validating beneficiary payout merchant policy rule",
		logger.String("merchantId", merchantId),
		logger.String("bankCode", bankCode),
		logger.String("accountNo", accountNo),
		logger.Float64("amount", amount),
	)

	var (
		processedValue = 0.0
		count          = int64(0)
	)
	limitResp := &disbursementModel.BeneficiaryPayoutLimitRuleLimit{}
	if rule != nil {
		limitResp.BeneficiaryPayoutLimitRuleConfig = *rule
	}

	cacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantId, bankCode, accountNo,
	)
	err := s.redisExt.HGetAllScan(ctx, cacheKey, limitResp)
	if err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Get beneficiary payout merchant policy rule limit from cache", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if errors.Is(err, redisExt.ErrNil) || limitResp.Count == 0 {
		// when cache not found or got 0 counter
		// re-validate by querying the database
		payoutLimit, err := s.GetBeneficiaryPayoutRuleLimitWithType(ctx, merchantId, bankCode, accountNo, rule, constant.DisbursementBeneficiaryLimitMerchantPolicy)
		if err != nil {
			s.logger.Error(ctx, "failed to re-validate beneficiary limit", logger.Error(err))
			return err
		}

		processedValue = payoutLimit.Processed
		count = int64(payoutLimit.Count)
	} else {
		processedValue, count, err = s.updateBeneficiaryQuota(ctx, cacheKey, amount)
		if err != nil {
			s.logger.Error(ctx, "failed to reduce beneficiary limit quota", logger.Error(err))
			return err
		}
	}

	if processedValue > limitResp.AmountThreshold || count > limitResp.Velocity {
		_, _ = s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", -amount)
		_, _ = s.redisExt.HIncrBy(ctx, cacheKey, "count", -1)

		err := pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions)

		s.sendBeneficiaryPayoutLimitAlert(ctx, disbursementModel.BeneficiaryPayoutLimitAlertRequest{
			TotalAmount:              processedValue,
			NumberOfTransaction:      count,
			BeneficiaryBankCode:      bankCode,
			BeneficiaryAccountNumber: accountNo,
			BeneficiaryAccountName:   accountName,
			MerchantID:               merchantId,
			AmountThreshold:          limitResp.AmountThreshold,
			CountThreshold:           limitResp.Velocity,
		})

		s.logger.Error(ctx, "[ValidateBeneficiaryPayoutMerchantPolicyRule] beneficiary limit exceeded", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "[ValidateBeneficiaryPayoutMerchantPolicyRule] Reduce beneficiary payout merchant policy custom rule limit", logger.Any("details", map[string]any{
		"merchantID": merchantId,
		"bankCode":   bankCode,
		"accountNo":  accountNo,
		"rule":       rule,
		"amount":     amount,
	}))

	return nil
}

func (s *DisbursementService) GetBeneficiaryPayoutRuleLimit(
	ctx context.Context, merchantId, bankCode, accountNo string, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error) {
	return s.GetBeneficiaryPayoutRuleLimitWithType(ctx, merchantId, bankCode, accountNo, rule, constant.DisbursementBeneficiaryLimitCustom)
}

func (s *DisbursementService) GetBeneficiaryPayoutRuleLimitWithType(
	ctx context.Context, merchantId, bankCode, accountNo string, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig, ruleType string,
) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetBeneficiaryPayoutRuleLimitWithType")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	var mutex redisExt.IMutexer
	if merchantId == "" {
		mutex = s.GetBeneficiaryPayoutDefaultRuleLimitLock("", bankCode, accountNo)
	} else if ruleType == constant.DisbursementBeneficiaryLimitMerchantPolicy {
		mutex = s.GetBeneficiaryPayoutMerchantPolicyRuleLimitLock(merchantId, bankCode, accountNo)
	} else {
		mutex = s.GetBeneficiaryPayoutCustomRuleLimitLock(merchantId, bankCode, accountNo)
	}

	if err := mutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "Failed lock process", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	defer func() {
		if _, err := mutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed unlock process", logger.Error(err))
		}
	}()

	result := &disbursementModel.BeneficiaryPayoutLimitRuleLimit{}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
	if merchantId != "" {
		if ruleType == constant.DisbursementBeneficiaryLimitMerchantPolicy {
			cacheKey = fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantId, bankCode, accountNo)
		} else {
			cacheKey = fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantId, bankCode, accountNo)
		}
	}

	if err := s.redisExt.HGetAllScan(ctx, cacheKey, result); err != nil {
		s.logger.Error(ctx, "Get data from cache with hgetall", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	}

	if result.Count > 0 && result.Processed > 0 {
		return result, nil
	}

	return s.calculateBeneficiaryPayoutRuleLimitWithType(ctx, merchantId, bankCode, accountNo, rule, ruleType)
}

func (s *DisbursementService) GetBeneficiaryPayoutCustomRuleLimitLock(
	merchantId, bankCode, accountNo string,
) redisExt.IMutexer {
	return s.redisExt.NewMutex(
		fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt+":lock", merchantId, bankCode, accountNo),
		redsync.WithTries(256),
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(50*time.Millisecond),
		redsync.WithFailFast(true),
	)
}

func (s *DisbursementService) GetBeneficiaryPayoutDefaultRuleLimitLock(
	merchantId, bankCode, accountNo string,
) redisExt.IMutexer {
	return s.redisExt.NewMutex(
		fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt+":lock", bankCode, accountNo),
		redsync.WithTries(256),
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(50*time.Millisecond),
		redsync.WithFailFast(true),
	)
}

func (s *DisbursementService) GetBeneficiaryPayoutMerchantPolicyRuleLimitLock(
	merchantId, bankCode, accountNo string,
) redisExt.IMutexer {
	return s.redisExt.NewMutex(
		fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt+":lock", merchantId, bankCode, accountNo),
		redsync.WithTries(256),
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(50*time.Millisecond),
		redsync.WithFailFast(true),
	)
}

func (s *DisbursementService) calculateBeneficiaryPayoutRuleLimit(
	ctx context.Context, merchantId, bankCode, accountNo string, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error) {
	return s.calculateBeneficiaryPayoutRuleLimitWithType(ctx, merchantId, bankCode, accountNo, rule, "custom")
}

func (s *DisbursementService) calculateBeneficiaryPayoutRuleLimitWithType(
	ctx context.Context, merchantId, bankCode, accountNo string, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig, ruleType string,
) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/calculateBeneficiaryPayoutRuleLimitWithType")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	wit := time.Now().In(local)

	// TODO: Implement rule timeframe later, need to finalize by product team
	startAt := time.Date(wit.Year(), wit.Month(), wit.Day(), 0, 0, 0, 0, local).In(time.UTC)
	endAt := time.Date(wit.Year(), wit.Month(), wit.Day(), 23, 59, 59, 999, local).In(time.UTC)

	result, err := s.disbursementRepo.GetBeneficiaryTransactionLimit(ctx, merchantId, bankCode, accountNo, startAt, endAt)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	if rule != nil {
		result.BeneficiaryPayoutLimitRuleConfig = *rule
	}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
	if merchantId != "" {
		if ruleType == constant.DisbursementBeneficiaryLimitMerchantPolicy {
			cacheKey = fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantId, bankCode, accountNo)
		} else {
			cacheKey = fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantId, bankCode, accountNo)
		}
	}

	_ = s.redisExt.HSet(ctx, cacheKey, "count", result.Count, "processed", result.Processed)
	_ = s.redisExt.Expire(ctx, cacheKey, endAt.Sub(time.Now().UTC()))

	return result, nil
}

func (s *DisbursementService) sendBeneficiaryPayoutLimitAlert(ctx context.Context, request disbursementModel.BeneficiaryPayoutLimitAlertRequest) {
	merchantName := ""
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant by id", logger.Error(err))
		return
	} else if merchant != nil {
		merchantName = merchant.Name
	}

	beneficiaryBankName := ""
	bank := bankDB.FindByCode(request.BeneficiaryBankCode)
	if bank != nil {
		beneficiaryBankName = bank.Name
	}

	slackMessage := &slackPb.PostWebhookCmd{
		URL:   s.config.SlackConfig.BeneficiaryPayoutLimitWebHookURL,
		Color: slackPb.Color_GOOD,
		Title: "Merchant got declined due to beneficiary limit",
		Fields: []*slackPb.AttachmentField{
			{Title: "Total Amount", Value: fmt.Sprintf("Rp %s", util.ConvertFloatToCurrency(request.TotalAmount)), Short: true},
			{Title: "Number of Transactions", Value: fmt.Sprintf("%d", request.NumberOfTransaction), Short: true},
			{Title: "Amount Threshold", Value: fmt.Sprintf("Rp %s", util.ConvertFloatToCurrency(request.AmountThreshold)), Short: true},
			{Title: "Count Threshold", Value: fmt.Sprintf("%d", request.CountThreshold), Short: true},
			{Title: "Beneficiary Account Number", Value: request.BeneficiaryAccountNumber, Short: true},
			{Title: "Beneficiary Account Name", Value: request.BeneficiaryAccountName, Short: true},
			{Title: "Beneficiary Bank Name", Value: beneficiaryBankName, Short: true},
			{Title: "Merchant ID", Value: request.MerchantID, Short: true},
			{Title: "Merchant Name", Value: merchantName, Short: true},
		},
	}
	rawSlackMessage, _ := proto.Marshal(slackMessage)
	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)
}

func (s *DisbursementService) updateBeneficiaryQuota(ctx context.Context, cacheKey string, amount float64) (float64, int64, error) {
	processedValue, err := s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", amount)
	if err != nil {
		s.logger.Error(ctx, "Incr processed value", logger.Error(err))
		return 0.0, 0, pkgErrs.New(response.HttpErrDatabase, err)

	}

	count, err := s.redisExt.HIncrBy(ctx, cacheKey, "count", 1)
	if err != nil {
		s.logger.Error(ctx, "Incr count value", logger.Error(err))
		return 0.0, 0, pkgErrs.New(response.HttpErrDatabase, err)

	}
	return processedValue, count, nil
}

// validateBeneficiaryLimit determines whether a given disbursement to a beneficiary
// is permitted according to merchant-level exemptions and beneficiary payout limit rules.
//
// Behavior:
//   - If the merchant is configured to be exempt from beneficiary rules, validation
//     is skipped and nil is returned.
//   - If either the disbursement or beneficiary is nil, an internal/data-not-found
//     error is logged and returned.
//   - The function inspects the beneficiary's metadata to obtain a custom payout
//     limit rule (if present).
//   - If the merchant is allowed to use the beneficiary's custom rule, the custom
//     validation path is invoked; otherwise the default validation path is invoked.
//   - Validation is delegated to the service's ValidateBeneficiaryPayoutCustomRule or
//     ValidateBeneficiaryPayoutDefaultRule methods. The disbursement's amount is
//     passed as a float64 (via InexactFloat64).
func (s *DisbursementService) validateBeneficiaryLimit(ctx context.Context, merchantID string, disbursement *disbursementModel.DisbursementWithTransaction, beneficary *beneficiaryAccountModel.Account) error {
	if _, isValid := s.IsMerchantAllowedExcludeBeneficiaryRules(ctx, merchantID, 0); isValid {
		return nil
	}

	if disbursement == nil || beneficary == nil {
		s.logger.Error(ctx, "beneficiary validation failure due to missing requirement")
		return pkgErrs.New(response.HttpErrInternal, constant.ErrDataNotFound)
	}

	beneficiaryLimitCustomRule := beneficary.MetadataObj.BeneficiaryPayoutLimitRule

	if s.IsMerchantAllowedToUseBeneficiaryCustomRule(ctx, merchantID, beneficiaryLimitCustomRule != nil) {
		return s.self.ValidateBeneficiaryPayoutCustomRule(
			ctx, merchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.BeneficiaryAccountName, disbursement.Amount.InexactFloat64(), beneficiaryLimitCustomRule,
		)
	}

	// Get merchant by ID to check for merchant-based custom rule
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant by id for beneficiary limit validation", logger.Error(err))
		// Continue with default rule if merchant fetch fails
	} else if merchant != nil {
		// Read merchant metadata to get beneficiary payout limit rule
		merchantMetadata, err := merchant.GetMetadata()
		if err != nil {
			s.logger.Error(ctx, "failed to parse merchant metadata", logger.Error(err))
			// Continue with default rule if merchant metadata fetch fails
		} else if merchantMetadata != nil && merchantMetadata.BeneficiaryPayoutLimitRule != nil {
			return s.self.ValidateBeneficiaryPayoutMerchantPolicyRule(
				ctx,
				merchantID,
				disbursement.BeneficiaryBankCode,
				disbursement.BeneficiaryAccountNo,
				disbursement.BeneficiaryAccountName,
				disbursement.Amount.InexactFloat64(),
				&disbursementModel.BeneficiaryPayoutLimitRuleConfig{
					Velocity:        merchantMetadata.BeneficiaryPayoutLimitRule.Velocity,
					Timeframe:       merchantMetadata.BeneficiaryPayoutLimitRule.Timeframe,
					AmountThreshold: merchantMetadata.BeneficiaryPayoutLimitRule.AmountThreshold,
				},
			)
		}
	}

	return s.self.ValidateBeneficiaryPayoutDefaultRule(
		ctx, merchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.BeneficiaryAccountName, disbursement.Amount.InexactFloat64(), beneficiaryLimitCustomRule,
	)
}
