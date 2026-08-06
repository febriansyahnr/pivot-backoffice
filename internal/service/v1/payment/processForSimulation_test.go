package paymentService

import (
	"context"
	"testing"
	"time"

	rabbitMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestProcessPaymentForSimulationByID(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)

	validReferenceID := "mock-reference-id"
	paymentMetadata := map[string]any{
		"snapCore": "OK",
	}

	validPayment := &paymentModel.Payment{
		PaymentMethodID: "mock-payment-method",
		MerchantID:      "mock-merchant-id",
		ReferenceID:     &validReferenceID,
		Metadata:        &paymentMetadata,
		Amount:          decimal.NewFromInt(10000),
		CreatedAt:       time.Now(),
	}

	testCases := []struct {
		name       string
		wantErr    bool
		status     string
		paidAmount *commonModel.Amount
		setupMock  func()
	}{
		{
			name:    "ERROR: Get payment by ID error repo",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get payment by ID not found",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Get payment method by ID not found",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrPaymentMethodNotFound)
			},
		},
		{
			name:    "ERROR: Get payment method by ID error repo",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: build QRIS metadata error on processQrisForSimulation",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.Payment{
					PaymentMethodID: "mock-payment-method",
					MerchantID:      "mock-merchant-id",
					ReferenceID:     &validReferenceID,
					Metadata:        nil,
				}, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

			},
		},
		{
			name:    "ERROR: paid amount is not match on processQrisForSimulation for dynamic QR type",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			paidAmount: &commonModel.Amount{
				Currency: "IDR",
				Value:    "5000",
			},
			setupMock: func() {
				qrMpmDynamicPayment := validPayment
				qrMpmDynamicPayment.Metadata = &map[string]any{
					"qrType":         constant.QrTypeDynamic,
					"validityPeriod": 10,
				}

				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(qrMpmDynamicPayment, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

			},
		},
		{
			name:    "ERROR: call to snapCore for QRIS payment",
			wantErr: true,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				qrMpmDynamicPayment := validPayment
				qrMpmDynamicPayment.Metadata = &map[string]any{
					"qrType":         constant.QrTypeDynamic,
					"validityPeriod": 10,
				}

				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(qrMpmDynamicPayment, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				snapCoreRepo.On("QrMpmPaymentSimulation", constant.ValueCtxMockType(),
					mock.AnythingOfType("*snapCoreModel.QrMpmPaymentSimulationRequest")).Once().Return(constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			name:    "SUCCESS: for QRIS payment",
			wantErr: false,
			status:  constant.ChargeStatusSuccess,
			setupMock: func() {
				qrMpmDynamicPayment := validPayment
				qrMpmDynamicPayment.Metadata = &map[string]any{
					"qrType":         constant.QrTypeDynamic,
					"validityPeriod": 10,
				}

				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(qrMpmDynamicPayment, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				snapCoreRepo.On("QrMpmPaymentSimulation", constant.ValueCtxMockType(),
					mock.AnythingOfType("*snapCoreModel.QrMpmPaymentSimulationRequest")).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, snapCoreRepo, nil, merchantRepo, paymentMethodRepo, nil)

			paidAmount := commonModel.Amount{
				Currency: "IDR",
				Value:    "10000.00",
			}
			if tc.paidAmount != nil {
				paidAmount = *tc.paidAmount
			}

			status := constant.ChargeStatusSuccess
			if tc.status != "" {
				status = tc.status
			}

			ctx := context.Background()
			err := paymentSvc.ProcessPaymentForSimulationByID(ctx, "mock-id", paidAmount, status)

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

func TestProcessPaymentForSimulationByIDExpiredAndProcessingStatus(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	rabbitMqExt := rabbitMock.NewRabbitMQExt(t)

	validReferenceID := "mock-reference-id"
	paymentMetadata := map[string]any{
		"qrType":         constant.QrTypeDynamic,
		"validityPeriod": 10,
	}

	validPayment := &paymentModel.Payment{
		UUID:            "mock-payment-uuid",
		PaymentMethodID: "mock-payment-method",
		MerchantID:      "mock-merchant-id",
		ReferenceID:     &validReferenceID,
		Metadata:        &paymentMetadata,
		Amount:          decimal.NewFromInt(10000),
		CreatedAt:       time.Now(),
	}

	testCases := []struct {
		name      string
		wantErr   bool
		status    string
		setupMock func()
	}{
		{
			name:    "SUCCESS: Publish expired status to RabbitMQ",
			wantErr: false,
			status:  constant.ChargeStatusExpired,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				rabbitMqExt.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*paymentModel.ExpiringPayment"),
					mock.AnythingOfType("time.Duration"),
				).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: Failed to publish expired status to RabbitMQ",
			wantErr: true,
			status:  constant.ChargeStatusExpired,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				rabbitMqExt.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*paymentModel.ExpiringPayment"),
					mock.AnythingOfType("time.Duration"),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: Publish processing status and continue to QRIS simulation",
			wantErr: false,
			status:  constant.ChargeStatusProcessing,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				rabbitMqExt.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*paymentModel.ExpiringPayment"),
					mock.AnythingOfType("time.Duration"),
				).Once().Return(nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				snapCoreRepo.On("QrMpmPaymentSimulation", constant.ValueCtxMockType(),
					mock.AnythingOfType("*snapCoreModel.QrMpmPaymentSimulationRequest")).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: Failed to publish processing status to RabbitMQ",
			wantErr: true,
			status:  constant.ChargeStatusProcessing,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				rabbitMqExt.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*paymentModel.ExpiringPayment"),
					mock.AnythingOfType("time.Duration"),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Publish processing success but QRIS simulation fails",
			wantErr: true,
			status:  constant.ChargeStatusProcessing,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(validPayment, nil)

				rabbitMqExt.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*paymentModel.ExpiringPayment"),
					mock.AnythingOfType("time.Duration"),
				).Once().Return(nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				snapCoreRepo.On("QrMpmPaymentSimulation", constant.ValueCtxMockType(),
					mock.AnythingOfType("*snapCoreModel.QrMpmPaymentSimulationRequest")).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(
				paymentRepo,
				logger,
				snapCoreRepo,
				nil,
				merchantRepo,
				paymentMethodRepo,
				nil,
				WithRabbitMQClient(rabbitMqExt),
			)

			paidAmount := commonModel.Amount{
				Currency: "IDR",
				Value:    "10000.00",
			}

			ctx := context.Background()
			err := paymentSvc.ProcessPaymentForSimulationByID(ctx, "mock-id", paidAmount, tc.status)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
			rabbitMqExt.AssertExpectations(t)
			if tc.status == constant.ChargeStatusProcessing {
				paymentMethodRepo.AssertExpectations(t)
				if !tc.wantErr || tc.name == "ERROR: Publish processing success but QRIS simulation fails" {
					snapCoreRepo.AssertExpectations(t)
				}
			}
		})
	}
}
