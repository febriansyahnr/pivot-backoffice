package feeService

import (
	"context"
	"fmt"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newIntCmd creates a redis.IntCmd with the given value (for mock returns).
func newIntCmd(val int64) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	cmd.SetVal(val)
	return cmd
}

// newIntCmdErr creates a redis.IntCmd with the given error (for mock returns).
func newIntCmdErr(err error) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	cmd.SetErr(err)
	return cmd
}

// newBoolCmd creates a redis.BoolCmd with the given value (for mock returns).
func newBoolCmd(val bool) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(context.Background())
	cmd.SetVal(val)
	return cmd
}

// newBoolCmdErr creates a redis.BoolCmd with the given error (for mock returns).
func newBoolCmdErr(err error) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(context.Background())
	cmd.SetErr(err)
	return cmd
}

// newStringCmd creates a redis.StringCmd with the given value (for mock returns).
func newStringCmd(val string) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(val)
	return cmd
}

// newStringCmdErr creates a redis.StringCmd with the given error (for mock returns).
func newStringCmdErr(err error) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetErr(err)
	return cmd
}

func TestResolveLadderTier(t *testing.T) {
	feeUUID := uuid.NewString()
	merchantId := uuid.NewString()
	now := time.Now().In(tz)
	redisKey := fmt.Sprintf(c.CacheKeyFmtMerchantFeeCounter, feeUUID, now.Format("2006-01"))

	tierConfigs := []merchantModel.FeeTieringConfig{
		{Tier: 1, Min: 0, Max: 5_000_000, AmountType: "AMOUNT", Amount: 10_000, TaxType: "NON_PKP"},
		{Tier: 2, Min: 5_000_001, Max: 10_000_000, AmountType: "AMOUNT", Amount: 7_500, TaxType: "NON_PKP"},
		{Tier: 3, Min: 10_000_001, Max: 999_999_999_999_999, AmountType: "AMOUNT", Amount: 5_000, TaxType: "NON_PKP"},
	}

	tpvType := c.TPVTieringType
	freqType := c.FrequencyTieringType

	tests := []struct {
		name        string
		merchantFee *merchantModel.MerchantFee
		request     *feeModel.GetFeeRequest
		setupMock   func(redisMock *redisMocks.IRedisExt, acctTrxRepo *repoMocks.IAccountTransactionRepository)
		wantTier    int
		wantNil     bool
		// Verify counter info returned in result
		wantRedisKey  string
		wantIncrement int64
	}{
		{
			name: "TPV tiering -- first transaction, no historical data",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &tpvType,
				TieringConfigsObj: tierConfigs,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "VIRTUAL_ACCOUNT",
				ReferenceAmount: 500_000,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, acctTrxRepo *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmdErr(redis.Nil))

				acctTrxRepo.On(
					"CalculatingMerchantTPVForLadderTiering",
					mock.Anything, merchantId, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
				).Once().Return(map[string]orchestratorModel.CalculatingMerchantTPVSummary{}, nil)
			},
			wantTier:      1, // cumulative = 0, falls in tier 1 (0 - 5,000,000)
			wantRedisKey:  redisKey,
			wantIncrement: 500_000,
		},
		{
			name: "TPV tiering -- key not found, mid-month config seeds Redis with historical TPV",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &tpvType,
				TieringConfigsObj: tierConfigs,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "VIRTUAL_ACCOUNT",
				ReferenceAmount: 1_000_000,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, acctTrxRepo *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmdErr(redis.Nil))

				acctTrxRepo.On(
					"CalculatingMerchantTPVForLadderTiering",
					mock.Anything, merchantId, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
				).Once().Return(map[string]orchestratorModel.CalculatingMerchantTPVSummary{
					"PAYMENT_VIRTUAL_ACCOUNT": {
						Type:    "PAYMENT",
						Channel: "VIRTUAL_ACCOUNT",
						Volume:  6_000_000,
					},
				}, nil)
				// Seed Redis with historical value
				redisMock.On("SetNX", mock.Anything, redisKey, int64(6_000_000), mock.AnythingOfType("time.Duration")).
					Return(newBoolCmd(true))
			},
			wantTier:      2, // cumulative = 6,000,000 -> tier 2 (5,000,001 - 10,000,000)
			wantRedisKey:  redisKey,
			wantIncrement: 1_000_000,
		},
		{
			name: "TPV tiering, Redis available -- subsequent transaction",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &tpvType,
				TieringConfigsObj: tierConfigs,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "VIRTUAL_ACCOUNT",
				ReferenceAmount: 2_000_000,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, _ *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmd("5500000"))
			},
			wantTier:      2, // current value = 5,500,000, falls in tier 2 (5,000,001 - 10,000,000)
			wantRedisKey:  redisKey,
			wantIncrement: 2_000_000,
		},
		{
			name: "FREQUENCY tiering, Redis available -- increment is always 1",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &freqType,
				TieringConfigsObj: tierConfigs,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "QRIS",
				ReferenceAmount: 100_000,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, _ *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmd("149"))
			},
			wantTier:      1, // current value = 149, falls in tier 1 (0 - 5,000,000)
			wantRedisKey:  redisKey,
			wantIncrement: 1,
		},
		{
			name: "Redis unavailable -- DB fallback with TPV",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &tpvType,
				TieringConfigsObj: tierConfigs,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "VIRTUAL_ACCOUNT",
				ReferenceAmount: 1_000_000,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, acctTrxRepo *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmdErr(c.ErrSomeErrorForUnitTest))

				acctTrxRepo.On(
					"CalculatingMerchantTPVForLadderTiering",
					mock.Anything, merchantId, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
				).Once().Return(map[string]orchestratorModel.CalculatingMerchantTPVSummary{
					"PAYMENT_VIRTUAL_ACCOUNT": {
						Type:    "PAYMENT",
						Channel: "VIRTUAL_ACCOUNT",
						Volume:  8_000_000,
					},
				}, nil)
			},
			wantTier:      2, // DB says volume = 8,000,000 -> tier 2
			wantRedisKey:  redisKey,
			wantIncrement: 1_000_000,
		},
		{
			name: "Empty TieringConfigsObj -- returns nil",
			merchantFee: &merchantModel.MerchantFee{
				UUID:              feeUUID,
				MerchantID:        merchantId,
				TieringType:       &tpvType,
				TieringConfigsObj: nil,
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				ReferenceAmount: 100_000,
			},
			setupMock: func(_ *redisMocks.IRedisExt, _ *repoMocks.IAccountTransactionRepository) {},
			wantNil:   true,
		},
		{
			name: "No matching tier -- defaults to tier 1",
			merchantFee: &merchantModel.MerchantFee{
				UUID:        feeUUID,
				MerchantID:  merchantId,
				TieringType: &tpvType,
				TieringConfigsObj: []merchantModel.FeeTieringConfig{
					{Tier: 1, Min: 100, Max: 200, AmountType: "AMOUNT", Amount: 5_000, TaxType: "NON_PKP"},
				},
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.TypePayment,
				PaymentMethod:   "VIRTUAL_ACCOUNT",
				ReferenceAmount: 50,
			},
			setupMock: func(redisMock *redisMocks.IRedisExt, acctTrxRepo *repoMocks.IAccountTransactionRepository) {
				redisMock.On("Get", mock.Anything, redisKey).
					Return(newStringCmdErr(redis.Nil))
				// DB returns 0 -- no historical transactions
				acctTrxRepo.On(
					"CalculatingMerchantTPVForLadderTiering",
					mock.Anything, merchantId, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
				).Once().Return(map[string]orchestratorModel.CalculatingMerchantTPVSummary{}, nil)
			},
			wantTier:      1,
			wantRedisKey:  redisKey,
			wantIncrement: 50,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			redisMock := redisMocks.NewIRedisExt(t)
			acctTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

			svc := &FeeService{
				logger:                 logger,
				accountTransactionRepo: acctTrxRepo,
				redis:                  redisMock,
			}

			if test.setupMock != nil {
				test.setupMock(redisMock, acctTrxRepo)
			}

			result := svc.resolveLadderTier(context.Background(), test.merchantFee, test.request)

			if test.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, test.wantTier, result.Tier.Tier)
				assert.Equal(t, test.wantRedisKey, result.RedisKey)
				assert.Equal(t, test.wantIncrement, result.Increment)
			}
		})
	}
}

func TestResolveLadderTier_DBFallbackFrequency(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	redisMock := redisMocks.NewIRedisExt(t)
	acctTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

	svc := &FeeService{
		logger:                 logger,
		accountTransactionRepo: acctTrxRepo,
		redis:                  redisMock,
	}

	feeUUID := uuid.NewString()
	merchantId := uuid.NewString()
	now := time.Now().In(tz)
	redisKey := fmt.Sprintf(c.CacheKeyFmtMerchantFeeCounter, feeUUID, now.Format("2006-01"))

	freqType := c.FrequencyTieringType
	tierConfigs := []merchantModel.FeeTieringConfig{
		{Tier: 1, Min: 0, Max: 100, AmountType: "AMOUNT", Amount: 10_000, TaxType: "NON_PKP"},
		{Tier: 2, Min: 101, Max: 999_999_999, AmountType: "AMOUNT", Amount: 5_000, TaxType: "NON_PKP"},
	}

	merchantFee := &merchantModel.MerchantFee{
		UUID:              feeUUID,
		MerchantID:        merchantId,
		TieringType:       &freqType,
		TieringConfigsObj: tierConfigs,
	}
	request := &feeModel.GetFeeRequest{
		MerchantID:      merchantId,
		Reference:       c.TypePayment,
		PaymentMethod:   "QRIS",
		ReferenceAmount: 50_000,
	}

	// Redis unavailable
	redisMock.On("Get", mock.Anything, redisKey).
		Return(newStringCmdErr(c.ErrSomeErrorForUnitTest))

	acctTrxRepo.On(
		"CalculatingMerchantTPVForLadderTiering",
		mock.Anything, merchantId, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
	).Once().Return(map[string]orchestratorModel.CalculatingMerchantTPVSummary{
		"PAYMENT_QR": {
			Type:      "PAYMENT",
			Channel:   "QR",
			Frequency: 150,
			Volume:    7_500_000,
		},
	}, nil)

	result := svc.resolveLadderTier(context.Background(), merchantFee, request)

	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Tier.Tier)         // frequency 150 -> tier 2 (101-999999999)
	assert.Equal(t, 5_000.0, result.Tier.Amount) // tier 2 amount
	assert.Equal(t, int64(1), result.Increment)  // frequency always increments by 1
}

func TestIncrementLadderCounter(t *testing.T) {
	tests := []struct {
		name      string
		redisKey  string
		increment int64
		setupMock func(redisMock *redisMocks.IRedisExt)
	}{
		{
			name:      "empty key -- no-op",
			redisKey:  "",
			increment: 1,
			setupMock: func(_ *redisMocks.IRedisExt) {},
		},
		{
			name:      "zero increment -- no-op",
			redisKey:  "some-key",
			increment: 0,
			setupMock: func(_ *redisMocks.IRedisExt) {},
		},
		{
			name:      "first transaction -- sets TTL",
			redisKey:  "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
			increment: 500_000,
			setupMock: func(redisMock *redisMocks.IRedisExt) {
				// result == increment means this is the first INCRBY (key was new)
				redisMock.On("IncrBy", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(500_000)).
					Return(newIntCmd(500_000))
				redisMock.On("Expire", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", mock.AnythingOfType("time.Duration")).
					Return(newBoolCmd(true))
			},
		},
		{
			name:      "subsequent transaction -- no TTL",
			redisKey:  "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
			increment: 1,
			setupMock: func(redisMock *redisMocks.IRedisExt) {
				// result (101) != increment (1) means key already existed
				redisMock.On("IncrBy", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(1)).
					Return(newIntCmd(101))
			},
		},
		{
			name:      "Redis IncrBy error -- logs and returns",
			redisKey:  "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
			increment: 1,
			setupMock: func(redisMock *redisMocks.IRedisExt) {
				redisMock.On("IncrBy", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(1)).
					Return(newIntCmdErr(c.ErrSomeErrorForUnitTest))
			},
		},
		{
			name:      "first transaction, Expire fails -- logs error but completes",
			redisKey:  "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
			increment: 100,
			setupMock: func(redisMock *redisMocks.IRedisExt) {
				redisMock.On("IncrBy", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(100)).
					Return(newIntCmd(100))
				redisMock.On("Expire", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", mock.AnythingOfType("time.Duration")).
					Return(newBoolCmdErr(c.ErrSomeErrorForUnitTest))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			redisMock := redisMocks.NewIRedisExt(t)

			svc := &FeeService{
				logger: logger,
				redis:  redisMock,
			}

			test.setupMock(redisMock)

			// Should not panic
			svc.IncrementLadderCounter(context.Background(), test.redisKey, test.increment)

			// Mock assertions are verified automatically by testify on cleanup
		})
	}
}

func TestSecondsUntilEndOfMonth(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Mid-month
	now := time.Date(2024, 3, 15, 12, 0, 0, 0, loc)
	ttl := secondsUntilEndOfMonth(now)
	expectedEnd := time.Date(2024, 4, 1, 0, 0, 0, 0, loc)
	assert.Equal(t, expectedEnd.Sub(now), ttl)

	// Last day of month
	lastDay := time.Date(2024, 3, 31, 23, 0, 0, 0, loc)
	ttl2 := secondsUntilEndOfMonth(lastDay)
	assert.True(t, ttl2 > 0)
	assert.True(t, ttl2 <= 1*time.Hour)

	// February (leap year)
	feb := time.Date(2024, 2, 28, 0, 0, 0, 0, loc)
	ttl3 := secondsUntilEndOfMonth(feb)
	expectedFebEnd := time.Date(2024, 3, 1, 0, 0, 0, 0, loc)
	assert.Equal(t, expectedFebEnd.Sub(feb), ttl3)
}
