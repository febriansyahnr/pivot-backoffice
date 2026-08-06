package refundProcessorService

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockRepos "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestRefundProcessorProcess(t *testing.T) {
	ctx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	testCases := []struct {
		name             string
		request          *refundModel.RefundProcessRequest
		mockSetup        func(*mockRepos.IRefundRepository, *mockServices.IOrchestratorService, *mockServices.IRefundService, *mockServices.IFeeService, *mockServices.ITransferService, *mockServices.IRefundProcessorService, *mockServices.IRefundProcessorService, *mockServices.IRefundProcessorService, *mockRedis.IRedisExt, *mockServices.IMerchantService)
		expectedError    string
		expectedResponse bool
	}{
		{
			name: "FAIL: Refund Not In Pending Status",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status: constant.RefundStatusSuccess,
				},
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
			},
			expectedError: constant.ErrRefundNotInPendingStatus.Error(),
		},
		{
			name: "FAIL: Redis Lock Error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status: constant.RefundStatusPending,
					UUID:   "refund-123",
				},
				RefundID: "refund-123",
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(false)
				redisCmd.SetErr(constant.ErrSomeErrorForUnitTest)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "FAIL: Redis Lock Already Acquired",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status: constant.RefundStatusPending,
					UUID:   "refund-123",
				},
				RefundID: "refund-123",
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(false)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
			},
			expectedError: constant.ErrRefundIsBeingProcessed.Error(),
		},
		{
			name: "FAIL: Begin Transaction Error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status: constant.RefundStatusPending,
					UUID:   "refund-123",
				},
				RefundID: "refund-123",
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "FAIL: Transfer Only Method Processor Error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodTransferOnly,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
				bankTransfer.On("Process", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "FAIL: Retrieve payment session charges",
			request: &refundModel.RefundProcessRequest{

				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, mock.Anything).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Unsettled Transaction Refund",
			request: &refundModel.RefundProcessRequest{

				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{
					SettlementStatus: sql.NullString{String: constant.StatusPending},
				}, nil)
				// card.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, mock.Anything).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				// feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Balance Equal To Refund Amount",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				card.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "FAIL: Card Process Error With Fallback",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				card.On("Process", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "FAIL: Card Facilitator Process Error With Fallback",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				card.On("Process", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "FAIL: Unsupported Payment Method Type",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        "INVALID_METHOD",
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "FAIL: Unsupported Refund Method",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:     constant.RefundStatusPending,
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     100000,
					Method:     constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistoryFailed).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
			},
			expectedError:    "",
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Card Refund Process",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:          constant.RefundStatusPending,
					UUID:            "refund-123",
					MerchantID:      "merchant-123",
					Amount:          100000,
					PaymentID:       "payment-123",
					PaymentChargeID: "charge-123",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				card.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Facilitator Channel Type Card Process",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:          constant.RefundStatusPending,
					UUID:            "refund-123",
					MerchantID:      "merchant-123",
					Amount:          100000,
					PaymentID:       "payment-123",
					PaymentChargeID: "charge-123",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
				PaymentMethodType:        paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				card.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Facilitator Channel Type QRIS Process",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:          constant.RefundStatusPending,
					UUID:            "refund-123",
					MerchantID:      "merchant-123",
					Amount:          100000,
					PaymentID:       "payment-123",
					PaymentChargeID: "charge-123",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
				PaymentMethodType:        paymentConstant.PAYMENT_METHOD_QRIS,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				qris.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: VA Transfer Method",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:          constant.RefundStatusPending,
					UUID:            "refund-123",
					MerchantID:      "merchant-123",
					Amount:          100000,
					PaymentID:       "payment-123",
					PaymentChargeID: "charge-123",
					DestinationType: constant.RefundDestinationTypeAccount,
					Method:          constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodVA,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()
				bankTransfer.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: With Parent Merchant And Fee Processing",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					Status:          constant.RefundStatusPending,
					UUID:            "refund-123",
					MerchantID:      "merchant-123",
					Amount:          100000,
					PaymentID:       "payment-123",
					PaymentChargeID: "charge-123",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
				},
				RefundID:                 "refund-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethodType:        constant.UnifiedPaymentMethodCard,
			},
			mockSetup: func(refundRepo *mockRepos.IRefundRepository, orchestratorSvc *mockServices.IOrchestratorService, refundSvc *mockServices.IRefundService, feeSvc *mockServices.IFeeService, transferSvc *mockServices.ITransferService, bankTransfer *mockServices.IRefundProcessorService, card *mockServices.IRefundProcessorService, qris *mockServices.IRefundProcessorService, redisMock *mockRedis.IRedisExt, merchantSvc *mockServices.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{Valid: true, String: "parent-merchant-123"},
				}, nil).Once()
				redisCmd := &redis.BoolCmd{}
				redisCmd.SetVal(true)
				redisMock.On("SetNX", mock.Anything, "backend-portal:refund-process:refund-123", true, 5*time.Minute).Return(redisCmd).Once()
				refundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				orchestratorSvc.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				card.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				refundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()

				// ChargeSubMerchantToMerchant
				feeSvc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{FinalAmount: 1000}, nil).Once()
				transferSvc.On("Transfer", mock.Anything, mock.Anything).Return(&transfer.Transfer{UUID: uuid.New()}, nil).Once()
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				// ChargeRefundFee
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(500.0, &feeModel.FeeMetadataObject{
					Type:        constant.TypeRefund,
					AmountType:  constant.MerchantFeeAmountType,
					FinalAmount: 500.0,
				}, nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()

				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				refundSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedResponse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRefundRepo := mockRepos.NewIRefundRepository(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			mockRefundSvc := mockServices.NewIRefundService(t)
			mockFeeSvc := mockServices.NewIFeeService(t)
			mockTransferSvc := mockServices.NewITransferService(t)
			mockBankTransfer := mockServices.NewIRefundProcessorService(t)
			mockCard := mockServices.NewIRefundProcessorService(t)
			mockQris := mockServices.NewIRefundProcessorService(t)
			mockRedis := mockRedis.NewIRedisExt(t)
			mockMerchantSvc := mockServices.NewIMerchantService(t)

			processor := &RefundProcessor{
				refundRepo:      mockRefundRepo,
				orchestratorSvc: mockOrchestratorSvc,
				refundSvc:       mockRefundSvc,
				feeSvc:          mockFeeSvc,
				transferSvc:     mockTransferSvc,
				bankTransfer:    mockBankTransfer,
				card:            mockCard,
				qris:            mockQris,
				redis:           mockRedis,
				logger:          pdkLog,
				merchantSvc:     mockMerchantSvc,
			}

			tc.mockSetup(mockRefundRepo, mockOrchestratorSvc, mockRefundSvc, mockFeeSvc, mockTransferSvc, mockBankTransfer, mockCard, mockQris, mockRedis, mockMerchantSvc)

			err := processor.Process(ctx, tc.request)

			if tc.expectedResponse {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}

			mockRefundRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			mockRefundSvc.AssertExpectations(t)
			mockFeeSvc.AssertExpectations(t)
			mockTransferSvc.AssertExpectations(t)
			mockBankTransfer.AssertExpectations(t)
			mockCard.AssertExpectations(t)
			mockQris.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}

func TestChargeSubMerchantToMerchant(t *testing.T) {
	var (
		transferID, _ = uuid.NewV7()
	)
	tests := []struct {
		name             string
		request          *refundModel.RefundProcessRequest
		setupMocks       func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService)
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name: "when failed to get fee, then should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID: "sub-merchant-id",
					Amount:     10000,
				},
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetTransactionFeeOnBehalf", constant.ValueCtxMockType(), mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when failed to transfer the balance, then should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:        "sub-merchant-id",
					ClientReferenceID: "rand-client-ref-id",
					Amount:            10000,
				},
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetTransactionFeeOnBehalf", constant.ValueCtxMockType(), mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{
					FinalAmount: 200,
				}, nil).Once()

				mockTransferSvc.On("Transfer", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when failed to update the transfer status, then should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:        "sub-merchant-id",
					ClientReferenceID: "rand-client-ref-id",
					Amount:            10000,
				},
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				var s *string
				mockFeeSvc.On("GetTransactionFeeOnBehalf", constant.ValueCtxMockType(), mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{
					FinalAmount: 200,
				}, nil).Once()

				mockTransferSvc.On("Transfer", mock.Anything, mock.Anything).Return(&transfer.Transfer{
					UUID: transferID,
				}, nil).Once()

				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transferID.String(), constant.StatusSuccess, s, s).
					Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "when everything is oke, then should return nil",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:        "sub-merchant-id",
					ClientReferenceID: "rand-client-ref-id",
					Amount:            10000,
				},
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				var s *string
				mockFeeSvc.On("GetTransactionFeeOnBehalf", constant.ValueCtxMockType(), mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{
					FinalAmount: 200,
				}, nil).Once()

				mockTransferSvc.On("Transfer", mock.Anything, mock.Anything).Return(&transfer.Transfer{
					UUID: transferID,
				}, nil).Once()

				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transferID.String(), constant.StatusSuccess, s, s).
					Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when fee metadata is zero, then should should not do transfer",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:        "sub-merchant-id",
					ClientReferenceID: "rand-client-ref-id",
					Amount:            10000,
				},
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetTransactionFeeOnBehalf", constant.ValueCtxMockType(), mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{
					FinalAmount: 0,
				}, nil).Once()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pdkLogger, _ := test.SetupLogger()

			mockFeeSvc := mockServices.NewIFeeService(t)
			mockTransferSvc := mockServices.NewITransferService(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)

			tt.setupMocks(mockFeeSvc, mockTransferSvc, mockOrchestratorSvc)

			processor := &RefundProcessor{
				feeSvc:          mockFeeSvc,
				transferSvc:     mockTransferSvc,
				orchestratorSvc: mockOrchestratorSvc,
				logger:          pdkLogger,
			}

			ctx := context.Background()
			// Add parent merchant ID to context
			ctx = context.WithValue(ctx, constant.CtxParentMerchantId, "parent-merchant-id")

			err := processor.ChargeSubMerchantToMerchant(ctx, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestCreditMDRToMerchant(t *testing.T) {
	var (
		feeLedgerID, _ = uuid.NewV7()
		parentID, _    = uuid.NewV7()
	)
	tests := []struct {
		name             string
		request          *refundModel.RefundProcessRequest
		setupMocks       func(mockOrchestratorSvc *mockServices.IOrchestratorService)
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name: "when failed to find fee ledger, then should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID: "merchant-id",
					PaymentID:  "payment-id",
				},
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when fee ledger is not found, then should return nil",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID: "merchant-id",
					PaymentID:  "payment-id",
				},
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(nil, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when percentage is zero, then should not create refund fee ledger",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID: "merchant-id",
					PaymentID:  "payment-id",
					Amount:     10000,
				},
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				metadataJSON, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{
					FeeMetadataObject: feeModel.FeeMetadataObject{
						AmountType:  "fixed",
						Percentage:  0,
						FinalAmount: 1000,
					},
				})

				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:     feeLedgerID,
						Currency: "IDR",
						AdditionalInfo: types.NullJSONText{
							Valid:    true,
							JSONText: metadataJSON,
						},
					}, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when refund of payment fee is greater than zero, then should create refund fee ledger",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:            "refund-id",
					MerchantID:      "merchant-id",
					PaymentID:       "payment-id",
					PaymentChargeID: "payment-charge-id",
					Amount:          10000,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				metadataJSON, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{
					FeeMetadataObject: feeModel.FeeMetadataObject{
						AmountType:  "percentage",
						Percentage:  2.5,
						FinalAmount: 250,
					},
				})

				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:     feeLedgerID,
						Currency: "IDR",
						AdditionalInfo: types.NullJSONText{
							Valid:    true,
							JSONText: metadataJSON,
						},
					}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFeeRefund &&
						req.Credit == 250.0 &&
						req.Status == constant.StatusSuccess
				})).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when failed to post account transaction, then should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "merchant-id",
					PaymentID:       "payment-id",
					PaymentChargeID: "payment-charge-id",
					Amount:          10000,
					UUID:            "refund-id",
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				metadataJSON, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{
					FeeMetadataObject: feeModel.FeeMetadataObject{
						AmountType:  "percentage",
						Percentage:  2.5,
						FinalAmount: 250,
					},
				})

				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:     feeLedgerID,
						Currency: "IDR",
						AdditionalInfo: types.NullJSONText{
							Valid:    true,
							JSONText: metadataJSON,
						},
					}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFeeRefund &&
						req.Credit == 250.0 &&
						req.Status == constant.StatusSuccess
				})).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "when parent merchant ID is present, then should use parent merchant ID",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "sub-merchant-id",
					PaymentID:       "payment-id",
					PaymentChargeID: "payment-charge-id",
					Amount:          10000,
					UUID:            "refund-id",
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService) {
				metadataJSON, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{
					FeeMetadataObject: feeModel.FeeMetadataObject{
						AmountType:  "percentage",
						Percentage:  2.5,
						FinalAmount: 250,
					},
				})

				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-id", constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:     feeLedgerID,
						Currency: "IDR",
						AdditionalInfo: types.NullJSONText{
							Valid:    true,
							JSONText: metadataJSON,
						},
					}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFeeRefund &&
						req.Credit == 250.0 &&
						req.Status == constant.StatusSuccess &&
						req.MerchantID.String() == parentID.String()
				})).Return(nil).Once()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pdkLogger, _ := test.SetupLogger()

			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			tt.setupMocks(mockOrchestratorSvc)

			processor := &RefundProcessor{
				orchestratorSvc: mockOrchestratorSvc,
				logger:          pdkLogger,
			}

			ctx := context.Background()
			if tt.name == "when parent merchant ID is present, then should use parent merchant ID" {
				// Add parent merchant ID to context
				ctx = context.WithValue(ctx, constant.CtxParentMerchantId, parentID.String())
			}

			err := processor.CreditMDRToMerchant(ctx, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestChargeRefundFee(t *testing.T) {
	var (
		parentID, _ = uuid.NewV7()
	)
	tests := []struct {
		name             string
		request          *refundModel.RefundProcessRequest
		setupMocks       func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService)
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name: "when fee calculation returns error, should log but continue processing",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "merchant-id",
					UUID:            "refund-id",
					DestinationType: constant.RefundDestinationTypeChannel,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(req *feeModel.GetFeeRequest) bool {
					return req.MerchantID == "merchant-id" &&
						req.Reference == constant.TypeRefund &&
						req.ReferenceType == constant.RefundDestinationTypeChannel
				})).Return(0.0, nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr: false,
		},
		{
			name: "when fee is zero, should not create fee ledger and return nil",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "merchant-id",
					UUID:            "refund-id",
					DestinationType: constant.RefundDestinationTypeChannel,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(req *feeModel.GetFeeRequest) bool {
					return req.MerchantID == "merchant-id" &&
						req.Reference == constant.TypeRefund &&
						req.ReferenceType == constant.RefundDestinationTypeChannel
				})).Return(0.0, &feeModel.FeeMetadataObject{}, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when fee is greater than zero, should create fee ledger and return nil",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "merchant-id",
					UUID:            "refund-id",
					DestinationType: constant.RefundDestinationTypeChannel,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(req *feeModel.GetFeeRequest) bool {
					return req.MerchantID == "merchant-id" &&
						req.Reference == constant.TypeRefund &&
						req.ReferenceType == constant.RefundDestinationTypeChannel
				})).Return(1000.0, &feeModel.FeeMetadataObject{
					Type:        constant.TypeRefund,
					AmountType:  constant.MerchantFeeAmountType,
					FinalAmount: 1000.0,
				}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFee &&
						req.Debit == 1000.0 &&
						req.Status == constant.StatusSuccess
				})).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "when post account transaction fails, should return error",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "merchant-id",
					UUID:            "refund-id",
					DestinationType: constant.RefundDestinationTypeChannel,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(req *feeModel.GetFeeRequest) bool {
					return req.MerchantID == "merchant-id" &&
						req.Reference == constant.TypeRefund &&
						req.ReferenceType == constant.RefundDestinationTypeChannel
				})).Return(1000.0, &feeModel.FeeMetadataObject{
					Type:        constant.TypeRefund,
					AmountType:  constant.MerchantFeeAmountType,
					FinalAmount: 1000.0,
				}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFee &&
						req.Debit == 1000.0 &&
						req.Status == constant.StatusSuccess
				})).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "when parent merchant ID is present, should use parent merchant ID",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					MerchantID:      "sub-merchant-id",
					UUID:            "refund-id",
					DestinationType: constant.RefundDestinationTypeChannel,
				},
				RefundID: "refund-id",
			},
			setupMocks: func(mockFeeSvc *mockServices.IFeeService, mockOrchestratorSvc *mockServices.IOrchestratorService) {
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(req *feeModel.GetFeeRequest) bool {
					return req.MerchantID == parentID.String() &&
						req.Reference == constant.TypeRefund &&
						req.ReferenceType == constant.RefundDestinationTypeChannel
				})).Return(1000.0, &feeModel.FeeMetadataObject{
					Type:        constant.TypeRefund,
					AmountType:  constant.MerchantFeeAmountType,
					FinalAmount: 1000.0,
				}, nil).Once()

				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.ReferenceID == "refund-id" &&
						req.Type == constant.TypeFee &&
						req.Debit == 1000.0 &&
						req.Status == constant.StatusSuccess &&
						req.MerchantID.String() == parentID.String()
				})).Return(nil).Once()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pdkLogger, _ := test.SetupLogger()

			mockFeeSvc := mockServices.NewIFeeService(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			tt.setupMocks(mockFeeSvc, mockOrchestratorSvc)

			processor := &RefundProcessor{
				feeSvc:          mockFeeSvc,
				orchestratorSvc: mockOrchestratorSvc,
				logger:          pdkLogger,
			}

			ctx := context.Background()
			if tt.name == "when parent merchant ID is present, should use parent merchant ID" {
				// Add parent merchant ID to context
				ctx = context.WithValue(ctx, constant.CtxParentMerchantId, parentID.String())
			}

			err := processor.ChargeRefundFee(ctx, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandlingReasonTypeAndDescFromTransfer(t *testing.T) {
	_, pdkLogger, _ := test.SetupLogger()

	processor := &RefundProcessor{
		logger: pdkLogger,
	}

	tests := []struct {
		name            string
		responseCode    string
		responseMessage string
		expectedReason  string
		expectedDesc    string
	}{
		{
			name:            "Insufficient fund pattern match",
			responseCode:    "4030014",
			responseMessage: "Insufficient balance",
			expectedReason:  constant.ReasonTypeInsufficientEscrowFund,
			expectedDesc:    "Insufficient balance",
		},
		{
			name:            "Inactive account pattern match",
			responseCode:    "4030018",
			responseMessage: "Account is inactive",
			expectedReason:  constant.ReasonTypeBeneficiaryAccountReason,
			expectedDesc:    constant.SnapCoreResponseInactiveAccountMessage,
		},
		{
			name:            "Dormant account pattern match",
			responseCode:    "4030009",
			responseMessage: "Account is dormant",
			expectedReason:  constant.ReasonTypeBeneficiaryAccountReason,
			expectedDesc:    constant.SnapCoreResponseDormantAccountMessage,
		},
		{
			name:            "Invalid account pattern match",
			responseCode:    "4040011",
			responseMessage: "Invalid account",
			expectedReason:  constant.ReasonTypeBeneficiaryAccountReason,
			expectedDesc:    constant.SnapCoreResponseInvalidAccountMessage,
		},
		{
			name:            "No pattern match returns empty",
			responseCode:    "999",
			responseMessage: "Unknown error",
			expectedReason:  "",
			expectedDesc:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasonType, reasonDesc := processor.handlingReasonTypeAndDescFromTransfer(tt.responseCode, tt.responseMessage)
			assert.Equal(t, tt.expectedReason, reasonType)
			assert.Equal(t, tt.expectedDesc, reasonDesc)
		})
	}
}

func TestGetTransactionByExternalID(t *testing.T) {
	ctx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	tests := []struct {
		name             string
		externalID       string
		setupMocks       func(*mockServices.IOrchestratorService, *mockRepos.IRefundRepository)
		expectedRefund   *refundModel.Refund
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name:       "FAIL: Error finding ledger by external ID",
			externalID: "external-123",
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundRepo *mockRepos.IRefundRepository) {
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:       "FAIL: Error finding refund by reference ID",
			externalID: "external-123",
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundRepo *mockRepos.IRefundRepository) {
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:       "FAIL: Refund not found",
			externalID: "external-123",
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundRepo *mockRepos.IRefundRepository) {
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(nil, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrRefundNotFound,
		},
		{
			name:       "SUCCESS: Refund found",
			externalID: "external-123",
			setupMocks: func(mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundRepo *mockRepos.IRefundRepository) {
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     10000,
				}, nil).Once()
			},
			expectedRefund: &refundModel.Refund{
				UUID:       "refund-123",
				MerchantID: "merchant-123",
				Amount:     10000,
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			mockRefundRepo := mockRepos.NewIRefundRepository(t)

			tt.setupMocks(mockOrchestratorSvc, mockRefundRepo)

			processor := &RefundProcessor{
				orchestratorSvc: mockOrchestratorSvc,
				refundRepo:      mockRefundRepo,
				logger:          pdkLog,
			}

			result, err := processor.getTransactionByExternalID(ctx, tt.externalID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRefund, result)
			}

			mockOrchestratorSvc.AssertExpectations(t)
			mockRefundRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateRefundToSuccess(t *testing.T) {
	ctx := context.Background()
	ctxTx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	tests := []struct {
		name             string
		request          *refundModel.RefundProcessRequest
		ctxValues        map[interface{}]interface{}
		setupMocks       func(*mockRepos.IRefundRepository, *mockServices.IOrchestratorService, *mockServices.IFeeService, *mockServices.ITransferService)
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name: "FAIL: Error updating refund data",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
				},
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS: Simple refund success update without parent merchant",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
				},
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "ledger-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "SUCCESS: Facilitator channel type skips fee processing",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
				},
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "ledger-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "SUCCESS: With parent merchant charges sub-merchant",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:              "refund-123",
					MerchantID:        "sub-merchant-123",
					Amount:            10000,
					ClientReferenceID: "client-ref-123",
				},
				RefundID:                 "refund-123",
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			ctxValues: map[interface{}]interface{}{
				constant.CtxParentMerchantId: "parent-merchant-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "ledger-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()

				// ChargeSubMerchantToMerchant mocks
				mockFeeSvc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{
					FinalAmount: 200,
				}, nil).Once()
				mockTransferSvc.On("Transfer", mock.Anything, mock.Anything).Return(&transfer.Transfer{
					UUID: uuid.New(),
				}, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()

				// ChargeRefundFee mocks
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(500.0, &feeModel.FeeMetadataObject{
					Type:        constant.TypeRefund,
					AmountType:  constant.MerchantFeeAmountType,
					FinalAmount: 500.0,
				}, nil).Once()
				mockOrchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "FAIL: ChargeSubMerchantToMerchant fails",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:              "refund-123",
					MerchantID:        "sub-merchant-123",
					Amount:            10000,
					ClientReferenceID: "client-ref-123",
				},
				RefundID:                 "refund-123",
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			ctxValues: map[interface{}]interface{}{
				constant.CtxParentMerchantId: "parent-merchant-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "ledger-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()

				// ChargeSubMerchantToMerchant fails
				mockFeeSvc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "FAIL: ChargeRefundFee fails",
			request: &refundModel.RefundProcessRequest{
				Refund: &refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
				},
				RefundLedgerReferenceID:  "ledger-123",
				PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockFeeSvc *mockServices.IFeeService, mockTransferSvc *mockServices.ITransferService) {
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "ledger-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()

				// ChargeRefundFee fails
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr: false, // Error is logged but not returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRefundRepo := mockRepos.NewIRefundRepository(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			mockFeeSvc := mockServices.NewIFeeService(t)
			mockTransferSvc := mockServices.NewITransferService(t)

			mockRefundSvc := mockServices.NewIRefundService(t)

			tt.setupMocks(mockRefundRepo, mockOrchestratorSvc, mockFeeSvc, mockTransferSvc)

			// Add mock for RecordRefundStatusHistory for cases that reach the status update
			// (all cases except the first one that fails at UpdateData)
			if tt.name != "FAIL: Error updating refund data" {
				mockRefundSvc.On("RecordRefundStatusHistory", mock.Anything, mock.Anything, constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return(nil).Once()
			}

			processor := &RefundProcessor{
				refundRepo:      mockRefundRepo,
				orchestratorSvc: mockOrchestratorSvc,
				refundSvc:       mockRefundSvc,
				feeSvc:          mockFeeSvc,
				transferSvc:     mockTransferSvc,
				logger:          pdkLog,
			}

			testCtx := ctx
			if tt.ctxValues != nil {
				for key, value := range tt.ctxValues {
					testCtx = context.WithValue(testCtx, key, value)
				}
			}

			err := processor.updateRefundToSuccess(ctxTx, testCtx, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, constant.RefundStatusSuccess, tt.request.Status)
			}

			mockRefundRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			mockFeeSvc.AssertExpectations(t)
			mockTransferSvc.AssertExpectations(t)
			mockRefundSvc.AssertExpectations(t)
		})
	}
}

func TestProcessUpdateBankTransferStatus(t *testing.T) {
	ctx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	refundLedgerID, _ := uuid.NewV7()
	paymentChargeID, _ := uuid.NewV7()
	paymentFeeID, _ := uuid.NewV7()

	tests := []struct {
		name             string
		request          *routingProcessorModel.BankTransferResponseData
		setupMocks       func(*mockRepos.IRefundRepository, *mockServices.IOrchestratorService, *mockServices.IRefundService, *mockServices.ISettlementService, *mockServices.IFeeService, *mockServices.IMerchantService)
		expectedErr      bool
		expectedErrValue error
	}{
		{
			name: "FAIL: Error getting transaction by external ID",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID: "external-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID fails
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "FAIL: Refund ledger not found",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID: "external-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID: "refund-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(nil, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound),
		},
		{
			name: "FAIL: Refund ledger not in pending status",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID: "external-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID: "refund-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found but not pending
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   refundLedgerID,
					Status: constant.StatusSuccess,
				}, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrRefundAlreadyProcessed),
		},
		{
			name: "FAIL: Payment charge not found",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID: "external-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID: "refund-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   refundLedgerID,
					Status: constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()

				// payment charge not found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(nil, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound),
		},
		{
			name: "FAIL: Payment fee not found",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID: "external-123",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID: "refund-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   refundLedgerID,
					Status: constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()

				// payment charge found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
				}, nil).Once()

				// payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound),
		},
		{
			name: "FAIL: Amount mismatch for non-Dana processor",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID:         "external-123",
				ProcessorReference: "OTHER_PROCESSOR",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000",
				},
				Status: constant.StatusSuccess,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:   "refund-123",
					Amount: 10000, // Different from request amount
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   refundLedgerID,
					Status: constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// payment charge found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
				}, nil).Once()

				// payment fee found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: paymentFeeID,
				}, nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()
			},
			expectedErr:      true,
			expectedErrValue: constant.ErrInvalidRequestPayload,
		},
		{
			name: "SUCCESS: Status pending - insufficient escrow fund",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID:         "external-123",
				ProcessorReference: "OTHER_PROCESSOR",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Status:          constant.StatusFailed,
				ResponseCode:    "4030014", // Insufficient fund pattern
				ResponseMessage: "Insufficient balance",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:   "refund-123",
					Amount: 10000,
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   refundLedgerID,
					Status: constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// payment charge found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
				}, nil).Once()

				// payment fee found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: paymentFeeID,
				}, nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "SUCCESS: Failed status update",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID:         "external-123",
				ProcessorReference: "OTHER_PROCESSOR",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Status:          constant.StatusFailed,
				ResponseCode:    "999",
				ResponseMessage: "General failure",
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:       "refund-123",
					Amount:     10000,
					MerchantID: "merchant-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        refundLedgerID,
					Status:      constant.StatusPending,
					ReferenceID: "refund-123",
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// payment charge found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
				}, nil).Once()

				// payment fee found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: paymentFeeID,
				}, nil).Once()

				// Begin transaction succeeds
				mockRefundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				// Update refund to failed
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "refund-123", constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()

				// Commit transaction
				mockRefundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()

				// Send callback
				mockRefundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "SUCCESS: Success status with settlement",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID:         "external-123",
				ProcessorReference: "OTHER_PROCESSOR",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Status: constant.StatusSuccess,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:       "refund-123",
					Amount:     10000,
					MerchantID: "merchant-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        refundLedgerID,
					Status:      constant.StatusPending,
					ReferenceID: "refund-123",
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// payment charge found with pending settlement
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
					SettlementStatus: sql.NullString{
						Valid:  true,
						String: constant.StatusPending,
					},
				}, nil).Once()

				// payment fee found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: paymentFeeID,
				}, nil).Once()

				// Begin transaction succeeds
				mockRefundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				// updateRefundToSuccess mocks (simplified)
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockRefundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "refund-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				// Commit transaction
				mockRefundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()

				// Force settlement
				mockSettlementSvc.On("ProcessSettlement", mock.Anything, mock.MatchedBy(func(req *settlementModel.ProcessSettlementRequest) bool {
					return req.MerchantID == "merchant-123" &&
						req.Type == constant.SettlementTransaction &&
						req.TransactionID == paymentChargeID.String() &&
						req.FeeTransactionID == paymentFeeID.String()
				})).Return(nil).Once()

				// Send callback
				mockRefundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(nil, nil).Once()
			},
			expectedErr: false,
		},
		{
			name: "SUCCESS: Dana processor skips amount validation",
			request: &routingProcessorModel.BankTransferResponseData{
				ExternalID:         "external-123",
				ProcessorReference: constant.DanaPGProcessor,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Status: constant.StatusSuccess,
			},
			setupMocks: func(mockRefundRepo *mockRepos.IRefundRepository, mockOrchestratorSvc *mockServices.IOrchestratorService, mockRefundSvc *mockServices.IRefundService, mockSettlementSvc *mockServices.ISettlementService, mockFeeSvc *mockServices.IFeeService, mockMerchantSvc *mockServices.IMerchantService) {
				// getTransactionByExternalID succeeds
				mockOrchestratorSvc.On("FindByID", mock.Anything, "external-123").Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "refund-123",
				}, nil).Once()
				mockRefundRepo.On("FindByID", mock.Anything, "refund-123").Return(&refundModel.Refund{
					UUID:       "refund-123",
					Amount:     10000, // Different from request but should be ignored for Dana
					MerchantID: "merchant-123",
				}, nil).Once()

				// find merchant
				mockMerchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil).Once()

				// refund ledger found and pending
				additionalInfo, _ := json.Marshal(orchestratorModel.MetadataRefund{
					PaymentChargeID: paymentChargeID.String(),
				})
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        refundLedgerID,
					Status:      constant.StatusPending,
					ReferenceID: "refund-123",
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfo,
					},
				}, nil).Once()

				// payment charge found
				mockOrchestratorSvc.On("FindByID", mock.Anything, paymentChargeID.String()).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        paymentChargeID,
					ReferenceID: "payment-123",
				}, nil).Once()

				// payment fee found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: paymentFeeID,
				}, nil).Once()

				// Begin transaction succeeds
				mockRefundRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()

				// updateRefundToSuccess mocks (simplified)
				mockRefundRepo.On("UpdateData", mock.Anything, mock.Anything).Return(nil).Once()
				mockRefundSvc.On("RecordRefundStatusHistory", mock.Anything, "refund-123", constant.StatusHistoryActorSystem, constant.RefundStatusHistorySuccess).Return().Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID", mock.Anything, "refund-123", constant.StatusSuccess, (*string)(nil), (*string)(nil)).Return(nil).Once()
				mockFeeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(0.0, nil, nil).Once()

				// Commit transaction
				mockRefundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()

				// Send callback
				mockRefundSvc.On("SendCallback", mock.Anything, "refund-123", "merchant-123").Return(nil).Once()

				// refund of payment fee not found
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "refund-123", constant.TypeFeeRefund).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Credit: 1000.00,
				}, nil).Once()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRefundRepo := mockRepos.NewIRefundRepository(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)
			mockRefundSvc := mockServices.NewIRefundService(t)
			mockSettlementSvc := mockServices.NewISettlementService(t)
			mockFeeSvc := mockServices.NewIFeeService(t)
			mockMerchantSvc := mockServices.NewIMerchantService(t)

			tt.setupMocks(mockRefundRepo, mockOrchestratorSvc, mockRefundSvc, mockSettlementSvc, mockFeeSvc, mockMerchantSvc)

			processor := &RefundProcessor{
				refundRepo:      mockRefundRepo,
				orchestratorSvc: mockOrchestratorSvc,
				refundSvc:       mockRefundSvc,
				settlementSvc:   mockSettlementSvc,
				logger:          pdkLog,
				feeSvc:          mockFeeSvc,
				merchantSvc:     mockMerchantSvc,
			}

			err := processor.ProcessUpdateBankTransferStatus(ctx, tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrValue != nil {
					assert.Equal(t, tt.expectedErrValue.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockRefundRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			mockRefundSvc.AssertExpectations(t)
			mockSettlementSvc.AssertExpectations(t)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
