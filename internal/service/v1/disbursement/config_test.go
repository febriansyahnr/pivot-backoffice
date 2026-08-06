package disbursementService_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionConfig(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	rdb, clientMock := redismock.NewClientMock()
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)

	service := New(
		&config.Config{}, logger, nil, disbursementRepo, nil, nil, WithRedisClient(redisExt.WrapRedisClient(rdb, nil)),
	)

	traceId := uuid.NewString()
	merchantId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
	cacheKey := fmt.Sprintf(c.DisbursementTransactionConfigFmt, merchantId)

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 35_000, MaxAmount: 85_000,
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *disbursementModel.TransactionConfig
	}{
		{
			name: "SUCCESS:Config in cache",
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.ExpectGet(cacheKey).SetVal(`{"minAmount": 15000, "maxAmount": 75000}`)
			},
			wantResult: &disbursementModel.TransactionConfig{
				MinAmount: 15_000, MaxAmount: 75_000,
			},
		},
		{
			name: "ERROR:Get config from cache",
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.ExpectGet(cacheKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId), // NOSONAR
		},
		{
			name: "ERROR:Get config from db",
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.ExpectGet(cacheKey).SetErr(redis.Nil)

				disbursementRepo.On("GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId), // NOSONAR
		},
		{
			name: "ERROR:Set config to cache",
			setupMock: func() {
				disbursementRepo.On("GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType()).Return(trxConfig, nil)

				clientMock.ExpectGet(cacheKey).SetErr(redis.Nil)
				clientMock.ExpectSet(cacheKey, trxConfig, 6*time.Hour).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId), // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redis.Nil)
				clientMock.ExpectSet(cacheKey, trxConfig, 6*time.Hour).SetVal("")
			},
			wantResult: trxConfig,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if result, err := service.GetTransactionConfig(ctx, merchantId); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetDailyTransactionLimit(t *testing.T) {
	// Redis Mocks
	mutex := redisMock.NewIMutexer(t)
	redisExt := redisMock.NewIRedisExt(t)

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)

	service := New(&config.Config{}, logger, merchantRepo, disbursementRepo, nil, nil, WithRedisClient(redisExt))

	merchantType := "MERCHANT"
	traceId := "7229d1a1-1012-4b1d-9935-53d696ade59a"
	merchantId := "c1ef2306-250e-4cf1-b737-5ae201438282"
	dailyTransactionLimit := disbursementModel.DailyTransactionLimitResponse{
		Limit:     util.ValueToPtr(1_000_000.00),
		Processed: 250_000,
		Remaining: 750_000,
	}
	errInternal := pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId))

	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
	cacheKey := fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, merchantType)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *disbursementModel.DailyTransactionLimitResponse
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: errInternal,
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Sub-Merchant data",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{String: c.KYCStatusNotRequired}}, nil)
			},
			wantErr: c.ErrForbiddenAccess,
		},
		{
			name: "ERROR:Lock context",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{String: c.KYCStatusApproved}}, nil)
				redisExt.On(
					"NewMutex", c.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(mutex)

				mutex.On("LockContext", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: errInternal,
		},
		{
			name: "ERROR:Get value from cache",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.Merchant{UUID: merchantId}, nil)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				redisExt.On(
					"HGetAllScan", c.ValueCtxMockType(), cacheKey, mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)

				mutex.On("UnlockContext", c.ValueCtxMockType()).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: errInternal,
		},
		{
			name: "SUCCESS:Get value from cache",
			setupMock: func() {
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)

				redisExt.On(
					"HGetAllScan", c.ValueCtxMockType(), cacheKey, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(2).(*disbursementModel.DailyTransactionLimitResponse)) = dailyTransactionLimit
				}).Return(nil)
			},
			wantResult: &dailyTransactionLimit,
		},
		{
			name: "ERROR:Get daily transaction limit",
			setupMock: func() {
				redisExt.On(
					"HGetAllScan", c.ValueCtxMockType(), cacheKey, mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(2).(*disbursementModel.DailyTransactionLimitResponse)) = disbursementModel.DailyTransactionLimitResponse{Limit: nil}
				}).Return(nil)

				disbursementRepo.On(
					"GetDailyTransactionLimit", c.ValueCtxMockType(), merchantId, merchantType,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: errInternal,
		},
		{
			name: "SUCCESS:Load and set",
			setupMock: func() {
				disbursementRepo.On(
					"GetDailyTransactionLimit", c.ValueCtxMockType(), merchantId, merchantType,
				).Return(&dailyTransactionLimit, nil)
				redisExt.On(
					"HSet", c.ValueCtxMockType(), cacheKey, "limit", dailyTransactionLimit.Limit, "processed", dailyTransactionLimit.Processed,
				).Return(&redis.IntCmd{})
				redisExt.On(
					"Expire", c.ValueCtxMockType(), cacheKey, c.DurationMockType(),
				).Return(&redis.BoolCmd{})
			},
			wantResult: &dailyTransactionLimit,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetDailyTransactionLimit(ctx, merchantId, merchantType)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
