package disbursementService

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetrySingle(t *testing.T) {
	type mocker struct {
		disbursementRepo    *repositoryMocks.IDisbursementRepository
		snapCoreRepo        *repositoryMocks.ISnapCoreRepository
		bankAccountRepo     *repositoryMocks.IBankAccountRepository
		orchestratorSvc     *serviceMocks.IOrchestratorService
		beneficiaryAccSvc   *serviceMocks.IBeneficiaryAccountService
		rmqExt              *rabbitMqMocks.RabbitMQExt
		forbiddenUsecaseSvc *serviceMocks.IMerchantForbiddenUseCaseService
		feeSvc              *serviceMocks.IFeeService
		routingProcessorSvc *serviceMocks.IRoutingProcessorService
		merchantRepo        *repositoryMocks.IMerchantRepository
		redisExt            *redisExtMocks.IRedisExt
		statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository
	}

	conf := config.Config{
		Environment: c.EnvironmentStaging,
		DisbursementConfig: config.DisbursementConfig{
			BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
				Amount:   10000000.00,
				Velocity: 100,
			},
		},
	}

	disbursementID := uuid.NewString()
	validRequest := &disbursementModel.RetrySingleRequest{
		DisbursementID: disbursementID,
		MerchantID:     uuid.NewString(),
	}

	feeDecimal := decimal.NewFromFloat(1000)

	validDisbursementWithTransaction := &disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
		UUID:                 validRequest.DisbursementID,
		Amount:               decimal.New(10000, 0),
		Fee:                  &feeDecimal,
		MerchantID:           validRequest.MerchantID,
		Status:               c.DisbursementStatusApproved,
		BeneficiaryBankCode:  "002",
		BeneficiaryAccountNo: "9999999666660001",
	}}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, "002", "9999999666660001")

	testCases := []struct {
		name       string
		wantErr    bool
		mocksSetup func(m *mocker)
		input      *disbursementModel.RetrySingleRequest
	}{
		{
			name:    "SUCCESS: Retry single",
			wantErr: false,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(&disbursementModel.ActionTransactionSummary{Total: 1, TotalAmount: 10000}, nil).
					Once()

				m.merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), validRequest.MerchantID,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   validRequest.MerchantID,
					DailyLimitMerchantType: "merchant",
				}, nil)

				m.merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "other-merchant-id",
					Name: "Test Merchant",
				}, nil)

				m.redisExt.On(
					"HGetScan", c.ValueCtxMockType(), c.StringMockType(), "limit", mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(3).(*float64) = 100_000.00
				}).Return(nil)

				m.redisExt.On(
					"HIncrByFloat", mock.Anything, c.StringMockType(), "processed", 10_000.00,
				).Times(1).Return(10_000.00, nil)

				m.redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				m.redisExt.On("HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10000.00).Return(2000.00, nil)
				m.redisExt.On("HIncrBy", c.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				m.beneficiaryAccSvc.On(
					"FindByBankCodeAndAccountNo",
					mock.Anything,
					CheckAccountReqMockType,
				).Return(&beneficiaryAccountModel.Account{}, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.rmqExt.On(
					"Publish",
					mock.AnythingOfType(c.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					c.PtrStringMockType(),
					mock.Anything,
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS: Retry single from Flip Processor",
			wantErr: false,
			mocksSetup: func(m *mocker) {
				flipProcessor := c.FlipPGProcessor
				disbursementFlipTransaction := validDisbursementWithTransaction
				disbursementFlipTransaction.ProcessorReferenceName = &flipProcessor

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(disbursementFlipTransaction, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(&disbursementModel.ActionTransactionSummary{Total: 1, TotalAmount: 10000}, nil).
					Once()

				m.merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), validRequest.MerchantID,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   validRequest.MerchantID,
					DailyLimitMerchantType: "merchant",
				}, nil)

				m.merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "other-merchant-id",
					Name: "Test Merchant",
				}, nil)

				m.redisExt.On(
					"HGetScan", c.ValueCtxMockType(), c.StringMockType(), "limit", mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(3).(*float64) = 100_000.00
				}).Return(nil)

				m.redisExt.On(
					"HIncrByFloat", mock.Anything, c.StringMockType(), "processed", 10_000.00,
				).Times(1).Return(10_000.00, nil)

				m.redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				m.redisExt.On("HIncrByFloat", c.ValueCtxMockType(), cacheKey, "processed", 10000.00).Return(2000.00, nil)
				m.redisExt.On("HIncrBy", c.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				m.beneficiaryAccSvc.On(
					"FindByBankCodeAndAccountNo",
					mock.Anything,
					CheckAccountReqMockType,
				).Return(&beneficiaryAccountModel.Account{}, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.rmqExt.On(
					"Publish",
					mock.AnythingOfType(c.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					c.PtrStringMockType(),
					mock.Anything,
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: FindByID got some error",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: FindByID not found",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Merchant not match",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					UUID:       validRequest.DisbursementID,
					MerchantID: "not valid",
					Fee:        &feeDecimal,
					Status:     c.DisbursementStatusApproved,
				}}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Status has not been approved",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					UUID:       validRequest.DisbursementID,
					MerchantID: validRequest.MerchantID,
					Fee:        &feeDecimal,
					Status:     c.DisbursementStatusWaiting,
				}}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Insufficient balance",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: UpdateReasonByIDs",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Failed to calculate total amount",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(nil, c.ErrSomeErrorForUnitTest).
					Once()

			},
			input: validRequest,
		},
		{
			name:    "ERROR: when transactionActoinSummary is nil, then should return error",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(nil, nil).
					Once()
			},
			input: validRequest,
		},
		{
			name:    "ERROR: when failed to validate the limit, then should return error",
			wantErr: true,
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(&disbursementModel.ActionTransactionSummary{Total: 1, TotalAmount: 10000}, nil).
					Once()

				m.merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), validRequest.MerchantID,
				).Return(nil, c.ErrSomeErrorForUnitTest)

			},
			input: validRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{
				disbursementRepo:    repositoryMocks.NewIDisbursementRepository(t),
				snapCoreRepo:        repositoryMocks.NewISnapCoreRepository(t),
				bankAccountRepo:     repositoryMocks.NewIBankAccountRepository(t),
				orchestratorSvc:     serviceMocks.NewIOrchestratorService(t),
				beneficiaryAccSvc:   serviceMocks.NewIBeneficiaryAccountService(t),
				rmqExt:              rabbitMqMocks.NewRabbitMQExt(t),
				forbiddenUsecaseSvc: serviceMocks.NewIMerchantForbiddenUseCaseService(t),
				feeSvc:              serviceMocks.NewIFeeService(t),
				routingProcessorSvc: serviceMocks.NewIRoutingProcessorService(t),
				merchantRepo:        repositoryMocks.NewIMerchantRepository(t),
				redisExt:            redisExtMocks.NewIRedisExt(t),
				statusHistoriesRepo: repositoryMocks.NewIStatusHistoriesRepository(t),
			}
			m.statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.mocksSetup(m)

			svc := New(
				&conf, pdkLoggerMock, m.merchantRepo, m.disbursementRepo, m.snapCoreRepo, m.bankAccountRepo,
				WithOrchestratorService(m.orchestratorSvc),
				WithBeneficiaryAccService(m.beneficiaryAccSvc),
				WithMerchantForbiddenUseCaseService(m.forbiddenUsecaseSvc),
				WithRedisClient(m.redisExt),
				WithFeeService(m.feeSvc),
				WithRoutingProcessorService(m.routingProcessorSvc),
				WithRabbitMQClient(m.rmqExt),
				WithStatusHistoriesRepository(m.statusHistoriesRepo),
			)

			ctx := context.Background()
			err := svc.RetrySingle(ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
