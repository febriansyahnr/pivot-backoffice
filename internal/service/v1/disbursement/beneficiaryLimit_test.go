package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/test"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestDecrBeneficiaryPayoutLimit(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redisExt := redisExtMocks.NewIRedisExt(t)

	disbursementService := &DisbursementService{
		logger:   logger,
		redisExt: redisExt,
	}

	merchantID := "ec7b478f-dd9d-4473-ba5b-7e1493c2c50e"
	bankCode := "002"
	accountNo := "9999999666660001"
	amount := 1000.00

	customCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo,
	)
	merchantPolicyCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantID, bankCode, accountNo,
	)
	defaultCacheKey := fmt.Sprintf(
		constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo,
	)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR: Checking custom cache key existence",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(0, errors.New("Redis error")))
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("Redis error")),
		},
		{
			name: "SUCCESS: Decrement custom rule",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(1, nil))
				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), customCacheKey, "processed", -amount).Once().Return(0.0, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), customCacheKey, "count", int64(-1)).Once().Return(int64(0), nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Decrement merchant policy rule when custom does not exist",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), merchantPolicyCacheKey).Once().Return(redis.NewIntResult(1, nil))
				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), merchantPolicyCacheKey, "processed", -amount).Once().Return(0.0, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), merchantPolicyCacheKey, "count", int64(-1)).Once().Return(int64(0), nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Decrement default rule when custom and merchant policy do not exist",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), merchantPolicyCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), defaultCacheKey).Once().Return(redis.NewIntResult(1, nil))
				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), defaultCacheKey, "processed", -amount).Once().Return(0.0, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), defaultCacheKey, "count", int64(-1)).Once().Return(int64(0), nil)
			},
			wantErr: nil,
		},
		{
			name: "ERROR: None of the rules exist",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), merchantPolicyCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), defaultCacheKey).Once().Return(redis.NewIntResult(0, nil))
			},
			wantErr: nil, // Function should return nil if no keys exist
		},
		{
			name: "ERROR: Checking merchant policy cache key existence",
			setupMock: func() {
				redisExt.On("Exists", constant.ValueCtxMockType(), customCacheKey).Once().Return(redis.NewIntResult(0, nil))
				redisExt.On("Exists", constant.ValueCtxMockType(), merchantPolicyCacheKey).Once().Return(redis.NewIntResult(0, errors.New("Redis error")))
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("Redis error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := disbursementService.DecrBeneficiaryPayoutLimit(context.Background(), merchantID, bankCode, accountNo, amount)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateBeneficiaryPayoutDefaultRule(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantId := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"
	accountName := "Dummy"
	amount := 1000.00

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)

	tests := []struct {
		name      string
		setupMock func() *DisbursementService
		wantErr   error
	}{
		{
			name: "error when getting from cache",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Return(errors.New("redis error")).Once()

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "cache miss - error from GetBeneficiaryPayoutRuleLimit",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				// First call from ValidateBeneficiaryPayoutDefaultRule
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 0
					}).
					Return(redis.Nil).Once()

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).Return(errors.New("not found"))

				return &DisbursementService{
					redisExt:         redisExt,
					logger:           logger,
					disbursementRepo: disbursementRepo,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "cache miss - successful rule retrieval",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				// First call from ValidateBeneficiaryPayoutDefaultRule
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 0
					}).
					Return(redis.Nil).Once()

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				// Mock returning existing data from cache instead of error
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 1
						limitResp.Processed = 100.00
					}).
					Return(nil)

				return &DisbursementService{
					redisExt:         redisExt,
					logger:           logger,
					disbursementRepo: disbursementRepo,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "success - cache hit with existing data, within limits",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "error from updateBeneficiaryQuota - HIncrByFloat fails",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(0.0, errors.New("redis error"))

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "error from updateBeneficiaryQuota - HIncrBy fails",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(1100.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(0), errors.New("redis error"))

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "limit exceeded - amount threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(6000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(5000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(5), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))
				// rabbitMqExt.On("Publish", constant.ValueCtxMockType(), "backend-portal.slack.post-webhook", (*string)(nil), mock.AnythingOfType("[]uint8")).Return(nil)

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   5000.00, // Low amount threshold to trigger limit
								Velocity: 100,
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
		{
			name: "limit exceeded - velocity threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(1000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(5), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))
				// rabbitMqExt.On("Publish", constant.ValueCtxMockType(), "backend-portal.slack.post-webhook", (*string)(nil), mock.AnythingOfType("[]uint8")).Return(nil)

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 5, // Low velocity threshold to trigger limit
							},
						},
					},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			err := service.ValidateBeneficiaryPayoutDefaultRule(ctx, merchantId, bankCode, accountNo, accountName, amount, rule)

			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestValidateBeneficiaryLimit(t *testing.T) {
	_, pdkLog, err := test.SetupLogger()
	if err != nil {
		panic(err)
	}
	defer pdkLog.Sync()

	ctx := context.Background()
	merchantID := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"
	accountName := "John Doe"
	amount := 1000.00

	tests := []struct {
		name         string
		merchantID   string
		disbursement *disbursementModel.DisbursementWithTransaction
		beneficiary  *beneficiaryAccountModel.Account
		setupMock    func() *DisbursementService
		wantErr      error
	}{
		{
			name:         "error - disbursement is nil",
			disbursement: nil,
			beneficiary: &beneficiaryAccountModel.Account{
				BeneficiaryBankCode:    bankCode,
				BeneficiaryAccountNo:   accountNo,
				BeneficiaryAccountName: accountName,
			},
			setupMock: func() *DisbursementService {
				return &DisbursementService{
					logger: pdkLog,
					config: &config.Config{},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrDataNotFound),
		},
		{
			name: "error - beneficiary is nil",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					BeneficiaryBankCode:    bankCode,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: accountName,
					Amount:                 decimal.NewFromFloat(amount),
				},
			},
			beneficiary: nil,
			setupMock: func() *DisbursementService {
				return &DisbursementService{
					logger: pdkLog,
					config: &config.Config{},
				}
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrDataNotFound),
		},
		{
			name: "success - merchant uses default rule",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					MerchantID:             "other-merchant-id",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: accountName,
					Amount:                 decimal.NewFromFloat(amount),
				},
			},
			beneficiary: &beneficiaryAccountModel.Account{
				BeneficiaryBankCode:    bankCode,
				BeneficiaryAccountNo:   accountNo,
				BeneficiaryAccountName: accountName,
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)

				// Mock merchant repo to return merchant without metadata
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{
					UUID: "other-merchant-id",
					Name: "Test Merchant",
				}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				service := &DisbursementService{
					logger:       pdkLog,
					redisExt:     redisExt,
					merchantRepo: merchantRepo,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
				service.self = service

				return service
			},
			wantErr: nil,
		},
		{
			name:       "success - merchant uses custom rule",
			merchantID: "custom-benef-limit-merchant-id",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					MerchantID:             "custom-benef-limit-merchant-id",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: accountName,
					Amount:                 decimal.NewFromFloat(amount),
				},
			},
			beneficiary: &beneficiaryAccountModel.Account{
				BeneficiaryBankCode:    bankCode,
				BeneficiaryAccountNo:   accountNo,
				BeneficiaryAccountName: accountName,
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
						Velocity:        10,
						AmountThreshold: 5000.00,
					},
				},
			},
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, "custom-benef-limit-merchant-id", bankCode, accountNo)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				service := &DisbursementService{
					logger:   pdkLog,
					redisExt: redisExt,
					config:   &config.Config{},
				}
				service.self = service

				return service
			},
			wantErr: nil,
		},
		{
			name:       "success - merchant uses merchant policy rule",
			merchantID: "merchant-policy-limit-merchant-id",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					MerchantID:             "merchant-policy-limit-merchant-id",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: accountName,
					Amount:                 decimal.NewFromFloat(amount),
				},
			},
			beneficiary: &beneficiaryAccountModel.Account{
				BeneficiaryBankCode:    bankCode,
				BeneficiaryAccountNo:   accountNo,
				BeneficiaryAccountName: accountName,
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil, // No beneficiary-level custom rule
				},
			},
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)

				// Mock merchant with metadata containing beneficiary payout limit rule
				merchantMetadataJSON := []byte(`{"kymNotes":"","beneficiaryPayoutLimitRule":{"velocity":15,"timeframe":"DAILY","amountThreshold":8000}}`)
				merchant := &merchantModel.Merchant{
					UUID: "merchant-policy-limit-merchant-id",
					Name: "Test Merchant",
					Metadata: types.NullJSONText{
						Valid:    true,
						JSONText: merchantMetadataJSON,
					},
				}

				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-policy-limit-merchant-id").Return(merchant, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, "merchant-policy-limit-merchant-id", bankCode, accountNo)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 15
						limitResp.AmountThreshold = 8000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2500.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				service := &DisbursementService{
					logger:       pdkLog,
					redisExt:     redisExt,
					merchantRepo: merchantRepo,
					config:       &config.Config{},
				}
				service.self = service

				return service
			},
			wantErr: nil,
		},
		{
			name:       "error - merchant repo fails when getting merchant metadata",
			merchantID: "merchant-policy-error-id",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					MerchantID:             "merchant-policy-error-id",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: accountName,
					Amount:                 decimal.NewFromFloat(amount),
				},
			},
			beneficiary: &beneficiaryAccountModel.Account{
				BeneficiaryBankCode:    bankCode,
				BeneficiaryAccountNo:   accountNo,
				BeneficiaryAccountName: accountName,
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)

				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-policy-error-id").Return(nil, errors.New("database error"))

				// Should fall back to default rule
				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				service := &DisbursementService{
					logger:       pdkLog,
					redisExt:     redisExt,
					merchantRepo: merchantRepo,
					config: &config.Config{
						DisbursementConfig: config.DisbursementConfig{
							BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
								Amount:   10000000.00,
								Velocity: 100,
							},
						},
					},
				}
				service.self = service

				return service
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			service := test.setupMock()
			mid := merchantID
			if test.merchantID != "" {
				mid = test.merchantID
			}

			err = service.validateBeneficiaryLimit(ctx, mid, test.disbursement, test.beneficiary)

			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestValidateBeneficiaryPayoutCustomRule(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantId := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"
	accountName := "Dummy"
	amount := 1000.00

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantId, bankCode, accountNo)

	tests := []struct {
		name      string
		setupMock func() *DisbursementService
		wantErr   error
	}{
		{
			name: "error when getting from cache",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Return(errors.New("redis error")).Once()

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "success - cache hit with existing data, within limits",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 10000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: nil,
		},
		{
			name: "error from updateBeneficiaryQuota - HIncrByFloat fails",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(0.0, errors.New("redis error"))

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "error from updateBeneficiaryQuota - HIncrBy fails",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(1100.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(0), errors.New("redis error"))

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "limit exceeded - amount threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(6000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(5000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(5), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
		{
			name: "limit exceeded - velocity threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 10000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(11), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(1000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(10), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			err := service.ValidateBeneficiaryPayoutCustomRule(ctx, merchantId, bankCode, accountNo, accountName, amount, rule)

			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestValidateBeneficiaryPayoutMerchantPolicyRule(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantId := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"
	accountName := "Dummy"
	amount := 1000.00

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantId, bankCode, accountNo)

	tests := []struct {
		name      string
		setupMock func() *DisbursementService
		wantErr   error
	}{
		{
			name: "error when getting from cache",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Return(errors.New("redis error")).Once()

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "cache miss - error from GetBeneficiaryPayoutRuleLimitWithType",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 0
					}).
					Return(redis.Nil).Once()

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).Return(errors.New("not found"))

				return &DisbursementService{
					redisExt:         redisExt,
					logger:           logger,
					disbursementRepo: disbursementRepo,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "success - cache hit with existing data, within limits",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 10000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: nil,
		},
		{
			name: "error from updateBeneficiaryQuota - HIncrByFloat fails",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(0.0, errors.New("redis error"))

				return &DisbursementService{
					redisExt: redisExt,
					logger:   logger,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, errors.New("redis error")),
		},
		{
			name: "limit exceeded - amount threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 5000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(6000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(5000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(5), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
		{
			name: "limit exceeded - velocity threshold",
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Velocity = 10
						limitResp.AmountThreshold = 10000.00
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", amount).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(11), nil)

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", -amount).Return(1000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(-1)).Return(int64(10), nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), merchantId).Return(nil, errors.New("merchant not found"))

				return &DisbursementService{
					redisExt:     redisExt,
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			err := service.ValidateBeneficiaryPayoutMerchantPolicyRule(ctx, merchantId, bankCode, accountNo, accountName, amount, rule)

			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestSendBeneficiaryPayoutLimitAlert(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantID := uuid.NewString()
	merchantName := "Test Merchant"
	bankCode := "002"
	accountNo := "9999999666660001"
	accountName := "John Doe"

	request := disbursementModel.BeneficiaryPayoutLimitAlertRequest{
		TotalAmount:              5000.00,
		NumberOfTransaction:      10,
		AmountThreshold:          10000.00,
		CountThreshold:           20,
		BeneficiaryAccountNumber: accountNo,
		BeneficiaryAccountName:   accountName,
		BeneficiaryBankCode:      bankCode,
		MerchantID:               merchantID,
	}

	tests := []struct {
		name      string
		setupMock func() *DisbursementService
	}{
		{
			name: "success - send alert with merchant found",
			setupMock: func() *DisbursementService {
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchantModel.Merchant{
					UUID: merchantID,
					Name: merchantName,
				}, nil)

				rabbitMqExt.On("Publish", mock.Anything, mock.AnythingOfType("string"), (*string)(nil), mock.AnythingOfType("[]uint8")).Return(nil)

				return &DisbursementService{
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
					config: &config.Config{
						SlackConfig: config.SlackConfig{
							BeneficiaryPayoutLimitWebHookURL: "https://hooks.slack.com/test",
						},
					},
				}
			},
		},
		{
			name: "success - send alert when merchant not found",
			setupMock: func() *DisbursementService {
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("merchant not found"))

				return &DisbursementService{
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
					config: &config.Config{
						SlackConfig: config.SlackConfig{
							BeneficiaryPayoutLimitWebHookURL: "https://hooks.slack.com/test",
						},
					},
				}
			},
		},
		{
			name: "success - send alert when merchant returns nil",
			setupMock: func() *DisbursementService {
				merchantRepo := repositoryMocks.NewIMerchantRepository(t)
				rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil)

				rabbitMqExt.On("Publish", mock.Anything, mock.AnythingOfType("string"), (*string)(nil), mock.AnythingOfType("[]uint8")).Return(nil)

				return &DisbursementService{
					logger:       logger,
					merchantRepo: merchantRepo,
					rabbitMqExt:  rabbitMqExt,
					config: &config.Config{
						SlackConfig: config.SlackConfig{
							BeneficiaryPayoutLimitWebHookURL: "https://hooks.slack.com/test",
						},
					},
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			// Function doesn't return error, just call it
			service.sendBeneficiaryPayoutLimitAlert(ctx, request)

			// Verify all expectations were met
			mock.AssertExpectationsForObjects(t, service.merchantRepo, service.rabbitMqExt)
		})
	}
}

func TestCalculateBeneficiaryPayoutRuleLimit(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantID := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	tests := []struct {
		name       string
		merchantID string
		rule       *disbursementModel.BeneficiaryPayoutLimitRuleConfig
		setupMock  func() *DisbursementService
		wantErr    error
	}{
		{
			name:       "error - GetBeneficiaryTransactionLimit returns error",
			merchantID: merchantID,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(nil, errors.New("database error"))

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name:       "success - calculate with custom rule and merchant ID",
			merchantID: merchantID,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     5,
						Processed: 2000.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 5, "processed", 2000.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - calculate with default rule (empty merchant ID)",
			merchantID: "",
			rule:       nil,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), "", bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     3,
						Processed: 1500.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 3, "processed", 1500.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - calculate with rule and verify rule is applied",
			merchantID: merchantID,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     7,
						Processed: 3500.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 7, "processed", 3500.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			result, err := service.calculateBeneficiaryPayoutRuleLimit(ctx, test.merchantID, bankCode, accountNo, test.rule)

			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify rule is applied if provided
				if test.rule != nil {
					assert.Equal(t, test.rule.Velocity, result.Velocity)
					assert.Equal(t, test.rule.AmountThreshold, result.AmountThreshold)
				}
			}
		})
	}
}

func TestCalculateBeneficiaryPayoutRuleLimitWithType(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantID := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	tests := []struct {
		name       string
		merchantID string
		ruleType   string
		rule       *disbursementModel.BeneficiaryPayoutLimitRuleConfig
		setupMock  func() *DisbursementService
		wantErr    error
	}{
		{
			name:       "success - calculate with merchant policy rule type",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitMerchantPolicy,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     5,
						Processed: 2000.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantID, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 5, "processed", 2000.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - calculate with custom rule type",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitCustom,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     3,
						Processed: 1500.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 3, "processed", 1500.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - calculate with default rule type (empty merchant ID)",
			merchantID: "",
			ruleType:   "",
			rule:       nil,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redisExt := redisExtMocks.NewIRedisExt(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), "", bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     7,
						Processed: 3500.00,
					}, nil)

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, bankCode, accountNo)
				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 7, "processed", 3500.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
					redisExt:         redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "error - GetBeneficiaryTransactionLimit fails",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitMerchantPolicy,
			rule:       rule,
			setupMock: func() *DisbursementService {
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(nil, errors.New("database error"))

				return &DisbursementService{
					logger:           logger,
					disbursementRepo: disbursementRepo,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			result, err := service.calculateBeneficiaryPayoutRuleLimitWithType(ctx, test.merchantID, bankCode, accountNo, test.rule, test.ruleType)

			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify rule is applied if provided
				if test.rule != nil {
					assert.Equal(t, test.rule.Velocity, result.Velocity)
					assert.Equal(t, test.rule.AmountThreshold, result.AmountThreshold)
				}
			}
		})
	}
}

func TestGetBeneficiaryPayoutRuleLimitWithType(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	merchantID := uuid.NewString()
	bankCode := "002"
	accountNo := "9999999666660001"

	rule := &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
		Velocity:        10,
		AmountThreshold: 5000.00,
	}

	tests := []struct {
		name       string
		merchantID string
		ruleType   string
		rule       *disbursementModel.BeneficiaryPayoutLimitRuleConfig
		setupMock  func() *DisbursementService
		wantErr    error
	}{
		{
			name:       "success - get from cache with merchant policy rule type",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitMerchantPolicy,
			rule:       rule,
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantID, bankCode, accountNo)

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
						limitResp.Processed = 2000.00
					}).
					Return(nil)

				return &DisbursementService{
					logger:   logger,
					redisExt: redisExt,
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - cache miss, recalculate with merchant policy type",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitMerchantPolicy,
			rule:       rule,
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutMerchantPolicyRuleLimitFmt, merchantID, bankCode, accountNo)

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				// Return empty count to trigger recalculation
				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 0
						limitResp.Processed = 0
					}).
					Return(nil)

				disbursementRepo.On("GetBeneficiaryTransactionLimit", constant.ValueCtxMockType(), merchantID, bankCode, accountNo, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(&disbursementModel.BeneficiaryPayoutLimitRuleLimit{
						Count:     3,
						Processed: 1500.00,
					}, nil)

				redisExt.On("HSet", constant.ValueCtxMockType(), cacheKey, "count", 3, "processed", 1500.00).Return(nil)
				redisExt.On("Expire", constant.ValueCtxMockType(), cacheKey, mock.AnythingOfType("time.Duration")).Return(nil)

				return &DisbursementService{
					logger:           logger,
					redisExt:         redisExt,
					disbursementRepo: disbursementRepo,
				}
			},
			wantErr: nil,
		},
		{
			name:       "error - lock fails",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitCustom,
			rule:       rule,
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(errors.New("lock error"))

				return &DisbursementService{
					logger:   logger,
					redisExt: redisExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name:       "error - HGetAllScan fails",
			merchantID: merchantID,
			ruleType:   constant.DisbursementBeneficiaryLimitCustom,
			rule:       rule,
			setupMock: func() *DisbursementService {
				redisExt := redisExtMocks.NewIRedisExt(t)
				mutex := redisExtMocks.NewIMutexer(t)
				redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

				cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutCustomRuleLimitFmt, merchantID, bankCode, accountNo)

				redisExt.On("NewMutex", constant.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)
				mutex.On("LockContext", constant.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", constant.ValueCtxMockType()).Return(false, nil)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).Return(errors.New("redis error"))

				return &DisbursementService{
					logger:   logger,
					redisExt: redisExt,
				}
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setupMock()

			result, err := service.GetBeneficiaryPayoutRuleLimitWithType(ctx, test.merchantID, bankCode, accountNo, test.rule, test.ruleType)

			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
