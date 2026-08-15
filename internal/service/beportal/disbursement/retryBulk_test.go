package disbursementService

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetryBulk(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
		DisbursementConfig: config.DisbursementConfig{
			BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
				Amount:   10000000.00,
				Velocity: 100,
			},
		},
	}
	validRequest := &disbursementModel.RetryBulkRequest{
		BulkDisbursementID: uuid.NewString(),
		MerchantID:         uuid.NewString(),
	}

	cacheKey := fmt.Sprintf(constant.BeneficiaryPayoutDefaultRuleLimitFmt, "002", "9999999666660001")

	testCases := []struct {
		name       string
		wantErr    bool
		errType    string
		mocksSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository,
			snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			orchestratorSvc *serviceMocks.IOrchestratorService,
			beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
			rmqExt *mockRabbitMq.RabbitMQExt,
			merchantRepo *repositoryMocks.IMerchantRepository,
			redisExt *redisExtMocks.IRedisExt,
		)
		input *disbursementModel.RetryBulkRequest
	}{
		{
			name:    "SUCCESS: Retry bulk",
			wantErr: false,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "other-merchant-id",
					Name: "Test Merchant",
				}, nil)

				feeAmount := decimal.NewFromFloat(2000.00)
				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:                 uuid.NewString(),
							BulkID:               &validRequest.BulkDisbursementID,
							MerchantID:           validRequest.MerchantID,
							Status:               constant.DisbursementStatusApproved,
							ReasonType:           &reasonType,
							Fee:                  &feeAmount,
							BeneficiaryBankCode:  "002",
							BeneficiaryAccountNo: "9999999666660001",
							Amount:               decimal.NewFromFloat32(1000.00),
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(&disbursementModel.ActionTransactionSummary{Total: 1, TotalAmount: 10000}, nil).
					Once()

				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), validRequest.MerchantID,
				).Return(&merchant.DisbursementMerchantConfig{
					DailyLimitMerchantId:   validRequest.MerchantID,
					DailyLimitMerchantType: "merchant",
				}, nil)

				redisExt.On(
					"HGetScan", c.ValueCtxMockType(), c.StringMockType(), "limit", mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(3).(*float64) = 100_000.00
				}).Return(nil)

				redisExt.On(
					"HIncrByFloat", mock.Anything, c.StringMockType(), "processed", 10_000.00,
				).Times(1).Return(10_000.00, nil)

				redisExt.On("HGetAllScan", mock.Anything, cacheKey, mock.Anything).
					Run(func(args mock.Arguments) {
						limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
						limitResp.Count = 5
					}).
					Return(nil).Once()

				redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), cacheKey, "processed", 1000.00).Return(2000.00, nil)
				redisExt.On("HIncrBy", constant.ValueCtxMockType(), cacheKey, "count", int64(1)).Return(int64(6), nil)

				rmqExt.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.Anything,
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				beneficiaryAccSvc.On(
					"FindByBankCodeAndAccountNo",
					mock.Anything,
					CheckAccountReqMockType,
				).Return(&beneficiaryAccountModel.Account{}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: FindBulkDisbursementByID",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: FindBulkDisbursementByID not found",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Merchant not match",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: "invalid merchant",
					Status:     constant.BulkDisbursementStatusPending,
				}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Status has not been approved",
			wantErr: true,
			errType: httpResponse.HttpErrUnprocessableContent,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.BulkDisbursementStatusUploading,
				}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: GetAllDisbursementByBulkID",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{}, constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Insufficient balance",
			wantErr: true,
			errType: httpResponse.HttpErrForbidden,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: UpdateReasonByIDs",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: UpdateBulkDisbursementStatusByID",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "when failed to get transaction summary, then should return error",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(nil, c.ErrSomeErrorForUnitTest).
					Once()
			},
			input: validRequest,
		},
		{
			name:    "when the summary was nil, then should return error",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(nil, nil).
					Once()

			},
			input: validRequest,
		},
		{
			name:    "when failed to get merchant config then should return error",
			wantErr: true,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				beneficiaryAccSvc *serviceMocks.IBeneficiaryAccountService,
				rmqExt *mockRabbitMq.RabbitMQExt,
				merchantRepo *repositoryMocks.IMerchantRepository,
				redisExt *redisExtMocks.IRedisExt,
			) {
				disbursementRepo.On(
					"FindBulkDisbursementByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursement{
					UUID:       validRequest.BulkDisbursementID,
					MerchantID: validRequest.MerchantID,
					Status:     constant.StatusPending,
				}, nil)

				reasonType := constant.DisbursementReasonTypeInsufficientBalance
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       uuid.NewString(),
							BulkID:     &validRequest.BulkDisbursementID,
							MerchantID: validRequest.MerchantID,
							Status:     constant.DisbursementStatusApproved,
							ReasonType: &reasonType,
						},
					},
				}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				disbursementRepo.On(
					"SumAmountByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType()).
					Return(&disbursementModel.ActionTransactionSummary{Total: 1, TotalAmount: 10000}, nil).
					Once()

				merchantRepo.On(
					"GetDisbursementMerchantConfig", c.ValueCtxMockType(), validRequest.MerchantID,
				).Return(nil, c.ErrSomeErrorForUnitTest)

			},
			input: validRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			mockRedisExt := redisExtMocks.NewIRedisExt(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.mocksSetup(disbursementRepoMock, snapCoreRepoMock, orchSvcMock, beneficiaryAccSvcMock, mockRmq, merchantRepo, mockRedisExt)

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, disbursementRepoMock, snapCoreRepoMock, nil,
				WithOrchestratorService(orchSvcMock), WithBeneficiaryAccService(beneficiaryAccSvcMock), WithRabbitMQClient(mockRmq), WithRedisClient(mockRedisExt),
				WithStatusHistoriesRepository(statusHistoriesRepo),
			)

			ctx := context.Background()
			err := svc.RetryBulk(ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errType != "" {
					extractedErrType, _ := pkgErrors.ExtractError(err)
					assert.Equal(t, tc.errType, extractedErrType)
				}
			} else {
				assert.NoError(t, err)
			}

			disbursementRepoMock.AssertExpectations(t)
			snapCoreRepoMock.AssertExpectations(t)
			orchSvcMock.AssertExpectations(t)
			beneficiaryAccSvcMock.AssertExpectations(t)
		})
	}
}
