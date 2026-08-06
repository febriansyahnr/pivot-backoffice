package feeService

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ladderTierResult holds the resolved tier and optional deferred counter info.
type ladderTierResult struct {
	Tier      *merchantModel.FeeTieringConfig
	RedisKey  string
	Increment int64
}

// resolveLadderTier resolves the matching tier from tiering_configs based on
// Falls back to DB query if Redis is unavailable.
func (s *FeeService) resolveLadderTier(ctx context.Context, merchantFee *merchantModel.MerchantFee, request *feeModel.GetFeeRequest) *ladderTierResult {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/resolveLadderTier")
	defer segment.End()

	tierConfigs := merchantFee.TieringConfigsObj
	if len(tierConfigs) == 0 {
		return nil
	}

	tieringType := ""
	if merchantFee.TieringType != nil {
		tieringType = *merchantFee.TieringType
	}

	// Determine the increment value based on tiering type
	var increment int64
	if tieringType == constant.TPVTieringType {
		increment = int64(math.Round(request.ReferenceAmount))
	} else {
		// FREQUENCY: increment by 1 per transaction
		increment = 1
	}

	now := time.Now().In(tz)
	redisKey := fmt.Sprintf(constant.CacheKeyFmtMerchantFeeCounter, merchantFee.UUID, now.Format("2006-01"))

	var cumulativeValue float64

	// Read current counter value BEFORE incrementing.
	// The fee is calculated based on cumulative volume/frequency prior to this transaction.
	currentVal, err := s.redis.Get(ctx, redisKey).Int64()
	if err != nil && !errors.Is(err, redisExt.ErrNil) {
		// Redis unavailable -- fall back to DB query
		s.logger.Error(ctx, "redis Get failed for ladder tiering, falling back to DB", logger.Error(err))
		cumulativeValue = s.ladderFallbackFromDB(ctx, merchantFee, request, tieringType, now)
	} else if errors.Is(err, redisExt.ErrNil) {
		// Key doesn't exist yet. Query DB for historical TPV/frequency for this month to initialize the counter.
		dbValue := s.ladderFallbackFromDB(ctx, merchantFee, request, tieringType, now)
		cumulativeValue = dbValue
		if dbValue > 0 {
			ttl := secondsUntilEndOfMonth(now)
			s.redis.SetNX(ctx, redisKey, int64(dbValue), ttl)
		}
	} else {
		cumulativeValue = float64(currentVal)
	}

	// Resolve tier based on cumulative value before this transaction
	resolvedTier := s.DetermineFeeTierLevel(ctx, cumulativeValue, tierConfigs)
	if resolvedTier == nil {
		resolvedTier = &tierConfigs[0]
	}

	return &ladderTierResult{
		Tier:      resolvedTier,
		RedisKey:  redisKey,
		Increment: increment,
	}
}

// IncrementLadderCounter increments the Redis ladder tiering counter.
func (s *FeeService) IncrementLadderCounter(ctx context.Context, redisKey string, increment int64) {
	if redisKey == "" || increment == 0 {
		return
	}

	result, err := s.redis.IncrBy(ctx, redisKey, increment).Result()
	if err != nil {
		s.logger.Error(ctx, "IncrementLadderCounter: redis IncrBy failed", logger.Error(err))
		return
	}

	if result == increment {
		now := time.Now().In(tz)
		ttl := secondsUntilEndOfMonth(now)
		if expErr := s.redis.Expire(ctx, redisKey, ttl).Err(); expErr != nil {
			s.logger.Error(ctx, "IncrementLadderCounter: failed to set expire", logger.Error(expErr))
		}
	}
}

// ladderFallbackFromDB queries the DB for current month TPV/frequency when Redis is unavailable.
func (s *FeeService) ladderFallbackFromDB(ctx context.Context, merchantFee *merchantModel.MerchantFee, request *feeModel.GetFeeRequest, tieringType string, now time.Time) float64 {
	if s.accountTransactionRepo == nil {
		s.logger.Error(ctx, "DB fallback for ladder tiering skipped: accountTransactionRepo is nil")
		return 0
	}

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, tz).UTC()
	nowUTC := now.UTC()

	tpvSummaries, err := s.accountTransactionRepo.CalculatingMerchantTPVForLadderTiering(ctx, request.MerchantID, startOfMonth, nowUTC)
	if err != nil {
		s.logger.Error(ctx, "DB fallback for ladder tiering failed", logger.Error(err))
		return 0
	}

	// Build lookup key to match the DB grouping (type + channel).
	// The account_transactions table stores channels differently from the fee
	// system's payment method names, so we map where needed:
	//   - Disbursement: always BANK_TRANSFER (regardless of specific bank)
	//   - QRIS: stored as "QR" in account_transactions, not "QRIS"
	//   - CREDIT_CARD: stored as "CARD" in account_transactions
	key := request.Reference
	if request.Reference == constant.ReferenceDisbursement {
		key += "_" + constant.ChannelBankTransfer
	} else if request.PaymentMethod != "" {
		pm := request.PaymentMethod
		switch pm {
		case constant.ChannelQris:
			pm = constant.UnifiedPaymentMethodQris
		case constant.ChannelCreditCard:
			pm = constant.UnifiedPaymentMethodCard
		}
		key += "_" + pm
	}

	summary := tpvSummaries[key]
	var result float64
	if tieringType == constant.FrequencyTieringType {
		result = summary.Frequency
	} else {
		result = summary.Volume
	}

	allKeys := make([]string, 0, len(tpvSummaries))
	for k := range tpvSummaries {
		allKeys = append(allKeys, k)
	}

	s.logger.Info(ctx, "DB fallback for ladder tiering: calculated cumulative value from DB", logger.Float64("cumulativeValue", result), logger.Any("allKeys", allKeys))

	return result
}

// secondsUntilEndOfMonth returns the duration from now until the last second
// of the current month in Asia/Jakarta timezone.
func secondsUntilEndOfMonth(now time.Time) time.Duration {
	firstOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return firstOfNextMonth.Sub(now)
}
