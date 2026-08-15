package disbursementService

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

type dailyLimitMerchant struct {
	redis  redisExt.IRedisExt
	logger logger.ILogger

	cacheKey     string
	referenceId  string
	merchantType string
	totalAmount  float64
}

func (s *DisbursementService) DecrDailyTransactionLimit(ctx context.Context, merchantId string, totalAmount float64) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/DecrDailyTransactionLimit")
	defer segment.End()

	merchantConfig, err := s.merchantRepo.GetDisbursementMerchantConfig(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Get disbursement merchant config", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	cacheKey := fmt.Sprintf(
		constant.DailyDisbursementTransactionConfigFmt, merchantConfig.DailyLimitMerchantId, merchantConfig.DailyLimitMerchantType,
	)
	if _, err = s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", -totalAmount); err != nil {
		s.logger.Error(ctx, "Decrement processed amount", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	s.logger.Info(ctx, "Restore daily transaction limit", logger.Any("details", map[string]any{
		"dailyLimitMerchantId":   merchantConfig.DailyLimitMerchantId,
		"dailyLimitMerchantType": merchantConfig.DailyLimitMerchantType,
		"totalAmount":            totalAmount,
	}))
	return nil
}

func (s *DisbursementService) ValidateDailyTransactionLimit(ctx context.Context, merchantId string, totalAmount float64) (service.ITransactionCloser, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateDailyTransactionLimit")
	defer segment.End()

	merchantConfig, err := s.merchantRepo.GetDisbursementMerchantConfig(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Get disbursement merchant config", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	dailyLimit := 0.0
	cacheKey := fmt.Sprintf(
		constant.DailyDisbursementTransactionConfigFmt, merchantConfig.DailyLimitMerchantId, merchantConfig.DailyLimitMerchantType,
	)
	if err := s.redisExt.HGetScan(ctx, cacheKey, "limit", &dailyLimit); err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Get daily transaction limit from cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if errors.Is(err, redisExt.ErrNil) {
		result, err := s.GetDailyTransactionLimit(ctx, merchantConfig.DailyLimitMerchantId, merchantConfig.DailyLimitMerchantType)
		if err != nil {
			return nil, err
		}

		dailyLimit = *result.Limit
	}

	processedValue, err := s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", totalAmount)
	if err != nil {
		s.logger.Error(ctx, "Incr processed value", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if processedValue > dailyLimit {
		_, _ = s.redisExt.HIncrByFloat(ctx, cacheKey, "processed", -totalAmount)
		remaining := dailyLimit - (processedValue - totalAmount)
		dailyLimitReachedErrString := fmt.Sprintf(constant.ErrMsgPayoutDailyLimitRemainingToday, util.ConvertFloatToCurrency(remaining))
		return nil, pkgErrs.New(response.HttpErrDailyLimitReached, errors.New(dailyLimitReachedErrString))
	}
	s.logger.Info(ctx, "Reduce daily transaction limits", logger.Any("details", map[string]any{
		"dailyLimitMerchantId":   merchantConfig.DailyLimitMerchantId,
		"dailyLimitMerchantType": merchantConfig.DailyLimitMerchantType,
		"totalAmount":            totalAmount,
	}))
	return &dailyLimitMerchant{
		redis:        s.redisExt,
		logger:       s.logger,
		cacheKey:     cacheKey,
		referenceId:  merchantConfig.DailyLimitMerchantId,
		merchantType: merchantConfig.DailyLimitMerchantType,
		totalAmount:  totalAmount,
	}, nil
}

func (s *dailyLimitMerchant) Close(ctx context.Context, status bool) (err error) {
	if status {
		return nil
	}

	_, err = s.redis.HIncrByFloat(ctx, s.cacheKey, "processed", -s.totalAmount)

	s.logger.Info(ctx, "Cancel daily transaction limits", logger.Any("details", map[string]any{
		"dailyLimitMerchantId":   s.referenceId,
		"dailyLimitMerchantType": s.merchantType,
		"totalAmount":            s.totalAmount,
	}))
	return err
}

// DeleteDailyTransactionLimit will resets the daily transaction limit for a given merchant in cache.
// it safe process because will reproduce new data when user access this data but not exist
func (s *DisbursementService) DeleteDailyTransactionLimit(ctx context.Context, merchantId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ResetDailyTransactionLimit")
	defer segment.End()

	merchantConfig, err := s.merchantRepo.GetDisbursementMerchantConfig(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Get disbursement merchant config", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	cacheKey := fmt.Sprintf(
		constant.DailyDisbursementTransactionConfigFmt, merchantConfig.DailyLimitMerchantId, merchantConfig.DailyLimitMerchantType,
	)

	err = s.redisExt.Del(ctx, cacheKey).Err()
	if err != nil {
		s.logger.Error(ctx, "Failed delete daily transaction limit", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	return nil
}
