package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
)

func (s *DisbursementService) GetTransactionConfig(ctx context.Context, merchantId string) (result *disbursementModel.TransactionConfig, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetTransactionConfig")
	defer segment.End()

	result = &disbursementModel.TransactionConfig{}
	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	configKey := fmt.Sprintf(constant.DisbursementTransactionConfigFmt, merchantId)

	if err = s.redisExt.Get(ctx, configKey).Scan(result); err == nil {
		return

	} else if !errors.Is(err, redis.Nil) {
		s.logger.Error(ctx, "getting disbursement transaction config from cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if result, err = s.disbursementRepo.GetTransactionConfig(ctx, merchantId); err != nil {
		s.logger.Error(ctx, "getting disbursement transaction config from db", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if err = s.redisExt.Set(ctx, configKey, result, 6*time.Hour).Err(); err != nil {
		s.logger.Error(ctx, "setting disbursement transaction config to cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	return
}

func (s *DisbursementService) GetDailyTransactionLimit(ctx context.Context, merchantId, merchantType string) (*disbursementModel.DailyTransactionLimitResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetDailyTransactionLimit")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Find merchant by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return nil, constant.ErrForbiddenAccess
	}

	mutex := s.GetDailyTransactionLimitLock(ctx, merchantId, merchantType)
	if err := mutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "Failed lock process", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	defer func() {
		if _, err := mutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed unlock process", logger.Error(err))
		}
	}()

	result := &disbursementModel.DailyTransactionLimitResponse{}
	cacheKey := fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, merchantId, merchantType)

	if err := s.redisExt.HGetAllScan(ctx, cacheKey, result); err != nil {
		s.logger.Error(ctx, "Get data from cache with hgetall", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if result.Limit != nil {
		result.Remaining = *result.Limit - result.Processed
		return result, nil
	}
	return s.calculateDailyTransactionLimit(ctx, merchantId, merchantType)
}

func (s *DisbursementService) GetDailyTransactionLimitLock(ctx context.Context, merchantId, merchantType string) redisExt.IMutexer {
	return s.redisExt.NewMutex(
		fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt+":lock", merchantId, merchantType),
		redsync.WithTries(256),
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(50*time.Millisecond),
		redsync.WithFailFast(true),
	)
}

func (s *DisbursementService) calculateDailyTransactionLimit(ctx context.Context, merchantId, merchantType string) (*disbursementModel.DailyTransactionLimitResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/calculateDailyTransactionLimit")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	result, err := s.disbursementRepo.GetDailyTransactionLimit(ctx, merchantId, merchantType)
	if err != nil {
		s.logger.Error(ctx, "Get daily transaction limit (db)", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	wit := time.Now().In(local)
	lastDatetime := time.Date(wit.Year(), wit.Month(), wit.Day(), 23, 59, 59, 999, local).In(time.UTC)

	cacheKey := fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, merchantId, merchantType)

	_ = s.redisExt.HSet(ctx, cacheKey, "limit", result.Limit, "processed", result.Processed)
	_ = s.redisExt.Expire(ctx, cacheKey, lastDatetime.Sub(time.Now().UTC()))

	result.Remaining = util.ValueOfPtr(result.Limit) - result.Processed
	return result, nil
}
