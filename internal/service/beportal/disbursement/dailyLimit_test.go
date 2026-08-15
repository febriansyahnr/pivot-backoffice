package disbursementService

import (
	"context"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDecrDailyTransactionLimit(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redisExt := redisExtMocks.NewIRedisExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := &DisbursementService{
		logger:       logger,
		merchantRepo: merchantRepo,
		redisExt:     redisExt,
	}

	merchantId := "ec7b478f-dd9d-4473-ba5b-7e1493c2c50e"
	cacheKey := fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, "merchant")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Get disbursement merchant config",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Decrement hash value",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId: merchantId, DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", -1_000.00,
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", -1_000.00,
				).Once().Return(0.0, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, service.DecrDailyTransactionLimit(context.Background(), merchantId, 1_000))
		})
	}

}

func TestValidateDailyTransactionLimit(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redisExt := redisExtMocks.NewIRedisExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := &DisbursementService{
		logger:       logger,
		redisExt:     redisExt,
		merchantRepo: merchantRepo,
	}

	ctx := context.Background()
	merchantId := "b5bf5553-fd2f-4565-b392-a31f32280eba"
	cacheKey := fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, "merchant")

	tests := []struct {
		name        string
		complate    *bool
		totalAmount float64
		setupMock   func()
		wantErr     error
	}{
		{
			name: "ERROR:Get disbursement merchant config",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Get daily transaction limit from cache",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   merchantId,
					DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"HGetScan", c.ValueCtxMockType(), cacheKey, "limit", mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Reload daily transaction limit",
			setupMock: func() {
				redisExt.On(
					"HGetScan", c.ValueCtxMockType(), cacheKey, "limit", mock.Anything,
				).Once().Return(redis.Nil)

				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, "")),
		},
		{
			name: "ERROR:Incr processed value",
			setupMock: func() {
				redisExt.On(
					"HGetScan", c.ValueCtxMockType(), cacheKey, "limit", mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(3).(*float64) = 10_000
				}).Return(nil)

				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10_000.00,
				).Once().Return(0.00, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Daily transaction limit reached",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   merchantId,
					DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"HGetScan", c.ValueCtxMockType(), cacheKey, "limit", mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(3).(*float64) = 10_000
				}).Return(nil)

				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10_000.00,
				).Times(1).Return(15_000.00, nil)
				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", -10_000.00,
				).Times(1).Return(5_000.00, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached,
				fmt.Errorf(c.ErrMsgPayoutDailyLimitRemainingToday, util.ConvertFloatToCurrency(5_000))),
		},
		{
			name:     "SUCCESS:With completed transaction",
			complate: util.ValueToPtr(true),
			setupMock: func() {
				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10_000.00,
				).Times(1).Return(10_000.00, nil)
			},
		},
		{
			name:     "SUCCESS:With uncompleted transaction",
			complate: util.ValueToPtr(false),
			setupMock: func() {
				redisExt.On(
					"HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10_000.00,
				).Times(1).Return(10_000.00, nil)
				redisExt.On(
					"HIncrByFloat", c.BackgroundCtxMockType(), cacheKey, "processed", -10_000.00,
				).Times(1).Return(0.00, nil)

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			if test.complate == nil {
				test.complate = util.ValueToPtr(false)
			}
			if test.totalAmount == 0 {
				test.totalAmount = 10_000
			}
			dailyLimit, err := service.ValidateDailyTransactionLimit(ctx, merchantId, test.totalAmount)
			assert.Equal(t, test.wantErr, err)

			if err == nil {
				require.NotNil(t, dailyLimit)
				dailyLimit.Close(ctx, *test.complate)
			}
		})
	}
}
func TestResetDailyTransactionLimit(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redisExt := redisExtMocks.NewIRedisExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := &DisbursementService{
		logger:       logger,
		redisExt:     redisExt,
		merchantRepo: merchantRepo,
	}

	merchantId := "d3b8f8f1-4c3b-4d8b-9f8b-1f8b8f8b8f8b"
	cacheKey := fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, "merchant")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Get disbursement merchant config",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Delete daily transaction limit from cache",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   merchantId,
					DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"Del", c.ValueCtxMockType(), cacheKey,
				).Once().Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest))
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   merchantId,
					DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"Del", c.ValueCtxMockType(), cacheKey,
				).Once().Return(redis.NewIntResult(1, nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.DeleteDailyTransactionLimit(context.Background(), merchantId)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
