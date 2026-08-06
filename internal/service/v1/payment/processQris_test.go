package paymentService

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	qrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreModelQr "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessQrisPayment(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
	qrisSvc := serviceMocks.NewIQrisService(t)
	feeSvc := serviceMocks.NewIFeeService(t)
	rmq := rabbitMqMocks.NewRabbitMQExt(t)
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
	accountTransactionRepo.On(
		"GetTransactionByReferenceIdAndProcessorId", mock.Anything, constant.StringMockType(), constant.StringMockType(),
	).Return(nil, nil)
	internalMethod := serviceMocks.NewIPaymentInternalDirectFunc(t)
	redis := redisExtMocks.NewIRedisExt(t)
	redisMutex := redisExtMocks.NewIMutexer(t)

	// Shared mocks for the distributed lock acquired when QrType is dynamic.
	// Uses Maybe() so cases that return before reaching the lock are not required to match.
	redis.On("NewMutex", mock.Anything, mock.Anything).Maybe().Return(redisMutex)
	redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
	redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

	buf := new(bytes.Buffer)
	logger := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))
	defer buf.Reset()

	validReferenceID := "mock-reference-id"
	validProcessorReferenceNo := "mock-processor-reference-no"
	paymentMetadata := map[string]any{
		"snapCore": &snapCoreModelQr.GenerateQrMpmResponseData{
			Amount: commonModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    "100000.00",
			},
		},
		"qrMethodType": constant.QrMethodTypeMPM,
		"qrType":       constant.QrTypeDynamic,
	}
	feeAmount := decimal.NewFromFloat(2500)
	expiredAt := time.Now().Add(24 * time.Hour)
	validPayment := paymentModel.Payment{
		PaymentMethodID:          "mock-payment-method",
		MerchantID:               uuid.NewString(),
		ReferenceID:              &validReferenceID,
		Metadata:                 &paymentMetadata,
		Amount:                   decimal.NewFromInt(10000),
		Fee:                      &feeAmount,
		ProcessorReferenceNumber: &validProcessorReferenceNo,
		ExpiredAt:                &expiredAt,
		Status:                   paymentConstant.QrisStatusPending,
	}
	validPaymentExpired := paymentModel.Payment{
		PaymentMethodID:          "mock-payment-method",
		MerchantID:               uuid.NewString(),
		ReferenceID:              &validReferenceID,
		Metadata:                 &paymentMetadata,
		Amount:                   decimal.NewFromInt(10000),
		Fee:                      &feeAmount,
		ProcessorReferenceNumber: &validProcessorReferenceNo,
		ExpiredAt:                &expiredAt,
		Status:                   paymentConstant.QrisStatusExpired,
	}

	testCases := []struct {
		name          string
		wantErr       bool
		requestStatus string
		setupMock     func()
	}{
		{
			name:    "ERROR: get active payment error repo",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: get active payment not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: determine payment fee",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On(
					"GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything,
				).Once().Return(&paymentModel.Payment{
					Status: paymentConstant.QrisStatusPending,
				}, nil)
				internalMethod.On(
					"DeterminePaymentFee", mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
			},
		},
		{
			name:    "ERROR: payment metadata is nil on buildQrisMetadata",
			wantErr: true,
			setupMock: func() {
				internalMethod.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					Status:          paymentConstant.QrisStatusPending,
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata:        nil,
					TotalAmount:     decimal.NewFromInt(10000),
				}, nil)

			},
		},
		{
			name:    "ERROR: unhandled QR Method CPM",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					Status:          paymentConstant.QrisStatusPending,
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata: &map[string]any{
						"qrMethodType": constant.QrMethodTypeCPM,
						"feeOnBehalf":  map[string]interface{}{"parentMerchantId": "3fc96de8-f65e-4b16-90a1-e2a00d1bae29"},
					},
					TotalAmount: decimal.NewFromInt(10000),
				}, nil)

			},
		},
		{
			name:    "ERROR: Invalid QR type",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					Status:          paymentConstant.QrisStatusPending,
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata: &map[string]any{
						"qrType": "invalid",
					},
					TotalAmount: decimal.NewFromInt(10000),
				}, nil)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got BeginTransaction error",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:          "ERROR: payment nil, QRIS is expired",
			requestStatus: constantPayment.QrisStatusExpired,
			wantErr:       false,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:          "ERROR: Idempotent, QRIS is not pending",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata:        &paymentMetadata,
					Status:          paymentConstant.QrisStatusSuccess,
					TotalAmount:     decimal.NewFromInt(50000),
				}, nil)

			},
		},
		{
			name:          "ERROR: Process success QRIS but payment amount is not match",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata:        &paymentMetadata,
					Status:          paymentConstant.QrisStatusPending,
					TotalAmount:     decimal.NewFromInt(50000),
				}, nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got updatePaymentStatus error",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got postLedger error",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Return(nil, nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchantModel.Merchant{}, nil)

				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Return(&paymentModel.PaymentMethodWithPivot{ChannelType: constant.PaymentMethodChannelTypeAggregator}, nil)

				orchestratorSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).
					Once().Return(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got CommitTransaction error",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       true,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				orchestratorSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).
					Times(2).Return(nil)

				feeSvc.On("CalculateFee",
					mock.Anything, constant.PtrGetFeeRequestMockType(), constant.PtrFeeMetadataObjectMockType(),
				).Return(1_000.00, 0.00)

				paymentRepo.On("GetPaymentById", mock.Anything, constant.StringMockType()).
					Return(&paymentModel.Payment{}, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchantModel.Merchant{KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					}}, nil)

				paymentRepo.On("CommitTransaction", mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got findMerchant error",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				orchestratorSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).
					Twice().Return(nil)

				paymentRepo.On("CommitTransaction", mock.Anything).
					Once().Return(nil)

				qrisSvc.On("FindQrRegistrationByExternalID", mock.Anything, mock.AnythingOfType("string")).
					Once().Return(&qrisModel.Registration{}, assert.AnError)
			},
		},
		{
			name:          "ERROR: Process success QRIS but got error find merchant",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				orchestratorSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).
					Twice().Return(nil)

				paymentRepo.On("CommitTransaction", mock.Anything).
					Once().Return(nil)

				qrisSvc.On("FindQrRegistrationByExternalID", mock.Anything, mock.AnythingOfType("string")).
					Once().Return(&qrisModel.Registration{}, assert.AnError)

			},
		},
		{
			name:          "SUCCESS: Process success QRIS",
			requestStatus: constantPayment.QrisStatusSuccess,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType()).
					Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)

				orchestratorSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).
					Times(2).Return(nil)

				paymentRepo.On("CommitTransaction", mock.Anything).
					Once().Return(nil)

				qrisSvc.On("FindQrRegistrationByExternalID", mock.Anything, mock.AnythingOfType("string")).
					Once().Return(&qrisModel.Registration{}, nil)

				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:          "SUCCESS: Process failed QRIS",
			requestStatus: constantPayment.QrisStatusFailed,
			wantErr:       false,
			setupMock: func() {
				payment := validPaymentExpired
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)

				paymentRepo.On(
					"UpdatePaymentStatus", mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(),
				).Once().Return(nil)

				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				orchestratorSvc.On(
					"PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:          "SUCCESS: Process EXPIRED QRIS",
			requestStatus: constantPayment.QrisStatusExpired,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)
			},
		},
		{
			name:          "SUCCESS: Process PENDING QRIS",
			requestStatus: constantPayment.QrisStatusPending,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)
			},
		},
		{
			name:          "SUCCESS: Process EXPIRED QRIS but the payment already expired by other process",
			requestStatus: constantPayment.QrisStatusExpired,
			wantErr:       false,
			setupMock: func() {
				payment := validPayment
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber",
					mock.Anything, mock.Anything).
					Once().Return(&payment, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, nil, nil, merchantRepo, paymentMethodRepo, nil,
				WithOrchestratorService(orchestratorSvc),
				WithRabbitMQClient(rmq),
				WithFeeService(feeSvc),
				WithQrisService(qrisSvc),
				WithAccountTransactionRepository(accountTransactionRepo),
				WithInternalDirectFunc(internalMethod),
				WithPaymentMethodService(paymentMethodSvc),
				WithRedisClient(redis),
			)

			requestStatus := constantPayment.QrisStatusPending
			if tc.requestStatus != "" {
				requestStatus = tc.requestStatus
			}

			ctx := context.Background()
			err := paymentSvc.ProcessQrisPayment(ctx, &paymentModel.QrisPaymentNotificationRequest{
				Acquirer:    constant.BANK_ACQUIRER_BNC,
				Status:      requestStatus,
				ReferenceNo: "mock-reference-no",
				PaidAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000.00",
				},
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
		})
	}
}
