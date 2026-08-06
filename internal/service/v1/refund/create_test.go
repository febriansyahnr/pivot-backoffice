package refundService

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestRefundService_Create(t *testing.T) {
	ctx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	testCases := []struct {
		name              string
		request           *refundModel.CreateRefundRequest
		mockSetup         func(*mocks.IRefundRepository, *mocks.IPaymentRepository, *mocks.IAccountTransactionRepository, *mocks.IPaymentMethodRepository, *serviceMocks.IOrchestratorService, *rabbitMqMocks.RabbitMQExt, *redisMocks.IRedisExt, *redisMocks.IMutexer)
		expectedError     string
		expectedErrorType string
		expectedResponse  bool
	}{
		{
			name: "FAIL: Client Reference Exists Db Error",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError:     constant.ErrSomeErrorForUnitTest.Error(),
			expectedErrorType: httpResponse.HttpErrDatabase,
		},
		{
			name: "FAIL: Client Reference Already Exists",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(true, nil).Once()
			},
			expectedError:     constant.ErrClientReferenceIDAlreadyExist.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Error Get Payment",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError:     constant.ErrSomeErrorForUnitTest.Error(),
			expectedErrorType: httpResponse.HttpErrDatabase,
		},
		{
			name: "FAIL: Payment Not Found",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(nil, nil).Once()
			},
			expectedError:     constant.ErrPaymentNotFound.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Merchant Id Mismatch",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID: "different-merchant",
				}, nil).Once()
			},
			expectedError:     constant.ErrMerchantIsNotMatch.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Payment Charge Not Found",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				ChargeID:          "charge-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
				}, nil).Once()
				accountTrxRepo.On("FindByID", mock.Anything, "charge-123").Return(nil, nil).Once()
			},
			expectedError:     constant.ErrPaymentChargeNotFound.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Error Get Payment Charge",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError:     constant.ErrSomeErrorForUnitTest.Error(),
			expectedErrorType: httpResponse.HttpErrDatabase,
		},
		{
			name: "FAIL: Payment Session ID != Payment Charge Reference ID",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				ChargeID:          "charge-123",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
				}, nil).Once()
				accountTrxRepo.On("FindByID", mock.Anything, mock.Anything).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil).Once()
			},
			expectedError:     constant.ErrPaymentChargeNotFound.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Error Get Refund",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError:     pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrInternalServerForUser).Error(),
			expectedErrorType: httpResponse.HttpErrDatabase,
		},
		{
			name: "FAIL: Refund Already Processed",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(100000.00, nil).Once()
			},
			expectedError:     constant.ErrPaymentAlreadyRefunded.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Facilitator Channel Type & Method Transfer",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodTransferOnly,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundIncorrectRequestMethodForFacilitator.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Facilitator Channel Type & Non Card Payment Method",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundNotAllowedForPaymentMethodFacilitatorConfig.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: CRM Request & Non BNC QRIS Payment Method",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				IsCRMRequest:      true,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundNotAllowedForPaymentMethodFacilitatorConfig.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Facilitator Channel Type & Non BNC QRIS Payment Method",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeFacilitator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundNotAllowedForPaymentMethodFacilitatorConfig.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Refund Amount Request Not Float",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      false,
				Amount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "200000x",
				},
				Method: constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
			},
			expectedError:     "",
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Refund Amount Exceeds Charge",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      false,
				Amount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "200000",
				},
				Method: constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        chargeUUID,
					ReferenceID: "payment-123",
					Currency:    "IDR",
					Credit:      100000,
					Status:      constant.ChargeStatusSuccess,
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundAmountExceedPaymentCharge.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Payment Charge Not Settled",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.ChargeStatusProcessing,
					SettlementStatus: sql.NullString{String: constant.StatusPending, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
			},
			expectedError:     constant.ErrPaymentChargeNotSettled.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Full refund on credit card within 24 hours",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 chargeUUID,
					ReferenceID:          "payment-123",
					Currency:             "IDR",
					Credit:               100000,
					Status:               constant.StatusSuccess,
					SettlementStatus:     sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel:      sql.NullString{String: constant.PaymentMethodChannelTypeAggregator, Valid: true},
					TransactionTimestamp: time.Now(),
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundIsNotYetAvailable.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Same as credit amount refund on credit card within 24 hours",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      false,
				Amount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "100000",
				},
				Method: constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 chargeUUID,
					ReferenceID:          "payment-123",
					Currency:             "IDR",
					Credit:               100000,
					Status:               constant.StatusSuccess,
					SettlementStatus:     sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel:      sql.NullString{String: constant.PaymentMethodChannelTypeAggregator, Valid: true},
					TransactionTimestamp: time.Now(),
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundIsNotYetAvailable.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "FAIL: Partial refund on credit card within 24 hours",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      false,
				Amount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "90000",
				},
				Method: constant.RefundMethodAuto,
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
					TransactionTimestamp: time.Now(),
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
			},
			expectedError:     constant.ErrRefundPartialIsNotYetAvailable.Error(),
			expectedErrorType: httpResponse.HttpErrUnprocessableContent,
		},
		{
			name: "SUCCESS: Full refund on credit card after 24 hours",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 chargeUUID,
					ReferenceID:          "payment-123",
					Currency:             "IDR",
					Credit:               100000,
					Status:               constant.StatusSuccess,
					SettlementStatus:     sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel:      sql.NullString{String: constant.PaymentMethodChannelTypeAggregator, Valid: true},
					TransactionTimestamp: time.Now().Add(-25 * time.Hour), // More than 24 hours ago
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Partial refund on credit card after 24 hours",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      false,
				Amount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "90000",
				},
				Method:      constant.RefundMethodAuto,
				Reason:      "test refund",
				Description: "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 chargeUUID,
					ReferenceID:          "payment-123",
					Currency:             "IDR",
					Credit:               100000,
					Status:               constant.StatusSuccess,
					SettlementStatus:     sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel:      sql.NullString{String: constant.PaymentMethodChannelTypeAggregator, Valid: true},
					TransactionTimestamp: time.Now().Add(-25 * time.Hour), // More than 24 hours ago
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: CRM Refund Card Payment",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				IsCRMRequest:      true,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: CRM Refund QRIS BNC Payment",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				IsCRMRequest:      true,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Refund Card Payment",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Refund QRIS BNC Payment",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
		{
			name: "SUCCESS: Successful Refund Creation",
			request: &refundModel.CreateRefundRequest{
				ClientReferenceID: "ref-123",
				MerchantID:        "merchant-123",
				PaymentSessionID:  "payment-123",
				IsFullAmount:      true,
				Method:            constant.RefundMethodAuto,
				Reason:            "test refund",
				Description:       "test description",
			},
			mockSetup: func(refundRepo *mocks.IRefundRepository, paymentRepo *mocks.IPaymentRepository, accountTrxRepo *mocks.IAccountTransactionRepository, paymentMethodRepo *mocks.IPaymentMethodRepository, orchestratorSvc *serviceMocks.IOrchestratorService, rabbitMq *rabbitMqMocks.RabbitMQExt, redisExt *redisMocks.IRedisExt, redisMutex *redisMocks.IMutexer) {
				chargeUUID := uuid.New()
				metadata := &types.JSONText{}
				metadataJson, _ := json.Marshal(map[string]interface{}{})
				metadata.UnmarshalJSON(metadataJson)
				refundRepo.On("ExistsByClientReferenceAndMerchantID", mock.Anything, "ref-123", "merchant-123").Return(false, nil).Once()
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Return(&paymentModel.Payment{
					MerchantID:      "merchant-123",
					PaymentMethodID: "pm-123",
					Metadata:        &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:             chargeUUID,
					ReferenceID:      "payment-123",
					Currency:         "IDR",
					Credit:           100000,
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
					SettlementModel: sql.NullString{
						String: constant.PaymentMethodChannelTypeAggregator,
						Valid:  true,
					},
				}, nil).Once()
				refundRepo.On("GetTotalRefundedAmount", mock.Anything, mock.Anything).Return(0.00, nil).Once()
				accountTrxRepo.On("FindByReference", mock.Anything, "payment-123", constant.TypeFee).Return(nil, nil).Once()

				redisExt.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisMutex).Once()
				redisMutex.On("LockContext", mock.Anything).Return(nil).Once()
				redisMutex.On("UnlockContext", mock.Anything).Return(true, nil).Once()

				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, "merchant-123", constant.TypePayment).Return(100000.0, nil).Once()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil).Once()
				refundRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()
				refundRepo.On("GetRefundList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{{
						ID:               "refund-123",
						MerchantID:       "merchant-123",
						PaymentSessionID: "payment-123",
						ChargeID:         chargeUUID.String(),
						Status:           constant.RefundStatusPending,
					}},
				}, nil).Once()
			},
			expectedResponse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRefundRepo := mocks.NewIRefundRepository(t)
			mockPaymentRepo := mocks.NewIPaymentRepository(t)
			mockAccountTrxRepo := mocks.NewIAccountTransactionRepository(t)
			mockPaymentMethodRepo := mocks.NewIPaymentMethodRepository(t)
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
			mockRedis := redisMocks.NewIRedisExt(t)
			mockRedisMutex := redisMocks.NewIMutexer(t)

			service := &RefundService{
				refundRepo:             mockRefundRepo,
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTrxRepo,
				paymentMethodRepo:      mockPaymentMethodRepo,
				orchestratorSvc:        mockOrchestratorSvc,
				rabbitMqExt:            mockRabbitMq,
				logger:                 pdkLog,
				redis:                  mockRedis,
			}

			tc.mockSetup(mockRefundRepo, mockPaymentRepo, mockAccountTrxRepo, mockPaymentMethodRepo, mockOrchestratorSvc, mockRabbitMq, mockRedis, mockRedisMutex)

			result, err := service.Create(ctx, tc.request)

			if tc.expectedResponse {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "merchant-123", result.MerchantID)
				assert.Equal(t, "ref-123", result.ClientReferenceID)
				assert.Equal(t, "payment-123", result.PaymentSessionID)
				assert.Equal(t, constant.RefundStatusPending, result.Status)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.expectedErrorType != "" {
					errType, _ := pkgErr.ExtractError(err)
					assert.Equal(t, tc.expectedErrorType, errType)
				}

				if tc.expectedError != "" {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			}

			mockRefundRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockAccountTrxRepo.AssertExpectations(t)
			mockPaymentMethodRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			mockRabbitMq.AssertExpectations(t)
		})
	}
}

func TestRefundService_calculateAndInsertMDRFee(t *testing.T) {
	ctx := context.Background()
	_, pdkLog, _ := test.SetupLogger()

	testCases := []struct {
		name                string
		paymentFeeLedger    *orchestratorModel.AccountTransactionWithUseCase
		refundAmount        float64
		paymentSessionID    string
		feeChargeID         string
		refundObj           *refundModel.Refund
		mockSetup           func(*serviceMocks.IOrchestratorService)
		expectedError       string
		expectedErrorType   string
		expectNoTransaction bool
	}{
		{
			name:                "SUCCESS: Nil payment fee ledger returns no error",
			paymentFeeLedger:    nil,
			refundAmount:        100000,
			paymentSessionID:    "payment-123",
			feeChargeID:         "fee-123",
			refundObj:           &refundModel.Refund{UUID: "refund-123", MerchantID: "merchant-123"},
			mockSetup:           func(orchestratorSvc *serviceMocks.IOrchestratorService) {},
			expectNoTransaction: true,
		},
		{
			name: "SUCCESS: Zero fee percentage returns no error",
			paymentFeeLedger: &orchestratorModel.AccountTransactionWithUseCase{
				UUID:     uuid.New(),
				Currency: "IDR",
				AdditionalInfo: types.NullJSONText{
					Valid: true,
					JSONText: func() []byte {
						b, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{FeeMetadataObject: feeModel.FeeMetadataObject{AmountType: constant.MerchantFeeAmountType, Percentage: 2.5}})
						return b
					}(),
				},
			},
			refundAmount:        100000,
			paymentSessionID:    "payment-123",
			feeChargeID:         "fee-123",
			refundObj:           &refundModel.Refund{UUID: "refund-123", MerchantID: "merchant-123"},
			mockSetup:           func(orchestratorSvc *serviceMocks.IOrchestratorService) {},
			expectNoTransaction: true,
		},
		{
			name: "SUCCESS: Non-zero fee percentage with zero refund amount",
			paymentFeeLedger: &orchestratorModel.AccountTransactionWithUseCase{
				UUID:     uuid.New(),
				Currency: "IDR",
				AdditionalInfo: types.NullJSONText{
					Valid: true,
					JSONText: func() []byte {
						b, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{FeeMetadataObject: feeModel.FeeMetadataObject{AmountType: constant.MerchantFeePercentageType, Percentage: 0.0}})
						return b
					}(),
				},
			},
			refundAmount:        100000,
			paymentSessionID:    "payment-123",
			feeChargeID:         "fee-123",
			refundObj:           &refundModel.Refund{UUID: "refund-123", MerchantID: "merchant-123"},
			mockSetup:           func(orchestratorSvc *serviceMocks.IOrchestratorService) {},
			expectNoTransaction: true,
		},
		{
			name: "SUCCESS: Valid fee percentage creates refund transaction",
			paymentFeeLedger: &orchestratorModel.AccountTransactionWithUseCase{
				UUID:     uuid.New(),
				Currency: "IDR",
				AdditionalInfo: types.NullJSONText{
					Valid: true,
					JSONText: func() []byte {
						b, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{FeeMetadataObject: feeModel.FeeMetadataObject{AmountType: constant.MerchantFeePercentageType, Percentage: 2.5}})
						return b
					}(),
				},
			},
			refundAmount:     100000,
			paymentSessionID: "payment-123",
			feeChargeID:      "fee-123",
			refundObj:        &refundModel.Refund{UUID: "refund-123", MerchantID: "merchant-123"},
			mockSetup: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.Type == constant.TypeFeeRefund &&
						req.ReferenceID == "refund-123" &&
						req.Currency == "IDR" &&
						req.Credit == 2500 &&
						req.Status == constant.StatusPending &&
						req.Usecase == constant.TypePayment
				})).Return(nil).Once()
			},
		},
		{
			name: "FAIL: PostAccountTransaction returns error",
			paymentFeeLedger: &orchestratorModel.AccountTransactionWithUseCase{
				UUID:     uuid.New(),
				Currency: "IDR",
				AdditionalInfo: types.NullJSONText{
					Valid: true,
					JSONText: func() []byte {
						b, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{FeeMetadataObject: feeModel.FeeMetadataObject{AmountType: constant.MerchantFeePercentageType, Percentage: 2.5}})
						return b
					}(),
				},
			},
			refundAmount:      100000,
			paymentSessionID:  "payment-123",
			feeChargeID:       "fee-123",
			refundObj:         &refundModel.Refund{UUID: "refund-123", MerchantID: "merchant-123"},
			expectedError:     constant.ErrRefundPaymentProcess.Error(),
			expectedErrorType: httpResponse.HttpErrDatabase,
			mockSetup: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
		},
		{
			name: "SUCCESS: High fee percentage calculation",
			paymentFeeLedger: &orchestratorModel.AccountTransactionWithUseCase{
				UUID:     uuid.New(),
				Currency: "USD",
				AdditionalInfo: types.NullJSONText{
					Valid: true,
					JSONText: func() []byte {
						b, _ := json.Marshal(orchestratorModel.FeeTransactionMetadataObject{FeeMetadataObject: feeModel.FeeMetadataObject{AmountType: constant.MerchantFeePercentageType, Percentage: 10.0}})
						return b
					}(),
				},
			},
			refundAmount:     50000,
			paymentSessionID: "payment-456",
			feeChargeID:      "fee-456",
			refundObj:        &refundModel.Refund{UUID: "refund-456", MerchantID: "merchant-456"},
			mockSetup: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.Type == constant.TypeFeeRefund &&
						req.ReferenceID == "refund-456" &&
						req.Currency == "USD" &&
						req.Credit == 5000 &&
						req.Status == constant.StatusPending &&
						req.Usecase == constant.TypePayment
				})).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)

			service := &RefundService{
				orchestratorSvc: mockOrchestratorSvc,
				logger:          pdkLog,
			}

			tc.mockSetup(mockOrchestratorSvc)

			err := service.calculateAndInsertMDRFee(ctx, tc.paymentFeeLedger, tc.refundAmount, tc.paymentSessionID, tc.feeChargeID, tc.refundObj)

			if tc.expectNoTransaction {
				assert.NoError(t, err)
			} else if tc.expectedError != "" {
				assert.Error(t, err)
				if tc.expectedErrorType != "" {
					errType, _ := pkgErr.ExtractError(err)
					assert.Equal(t, tc.expectedErrorType, errType)
				}
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockOrchestratorSvc.AssertExpectations(t)
		})
	}
}
