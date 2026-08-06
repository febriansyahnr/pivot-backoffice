package creditcard_test

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/creditcard"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentNotification(t *testing.T) {
	conf := &config.Config{CreditcardConfig: config.CreditcardConfig{WebviewURL: "", ProcessorURL: ""}}

	now := time.Now()
	expiredAt := now.Add(constant.CreditCardPaymentExpired)

	referenceId := "some-reference-id"
	paymentUUID := uuid.New()
	merchantId := uuid.New()
	paymentMethodID := uuid.New()

	validRequest := creditcardModel.CardPaymentNotificationRequest{
		Event: "CREDIT_CARD_PAYMENTS_CALLBACK",
		Data: creditcardModel.PaymentNotificationDataRequest{
			PaymentUUID:           uuid.New(),
			ReferenceID:           referenceId,
			MerchantID:            merchantId,
			Amount:                decimal.NewFromFloat(10000),
			Currency:              "IDR",
			AcquirerTransactionID: "acquirer-transaction-id",
			PaymentStatus:         "SUCCESS",
			CardData:              &creditcardModel.CardDataRequest{First8Digit: "12345678", Last4Digit: "4321", CardType: "credit", CardBrand: "visa", CardIssuing: "bank", CountryCode: "ID", Fingerprint: "fingerprint"},
			Updated:               now,
		},
	}

	paymentMethodDB := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID:      paymentMethodID.String(),
			Type:      paymentMethodConstant.PAYMENT_METHOD_CREDIT_CARD,
			Category:  paymentMethodConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
			Name:      "Creditcard Payment",
			Acquirer:  "HARSYA",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	paymentDB := &paymentModel.Payment{
		UUID:        paymentUUID.String(),
		ReferenceID: &referenceId,
		Amount:      decimal.NewFromFloat(10000),
		Currency:    "IDR",
		PaymentURL:  "https://example.com/pay/existing",
		Status:      "PENDING",
		ExpiredAt:   &expiredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	creditCardDB := &paymentModel.Payment{
		ReferenceID: &referenceId,
		Metadata:    &map[string]any{"feeOnBehalf": map[string]any{"parentMerchantId": "3fc96de8-f65e-4b16-90a1-e2a00d1bae29"}},
	}

	merchantData := &merchantModel.Merchant{
		BusinessCountry: sql.NullString{String: "ID"},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(
			rmqExtMock *rabbitMqMocks.RabbitMQExt,
			paymentSvc *serviceMocks.IPaymentService,
			mockPaymentRepo *repositoryMocks.IPaymentRepository,
			paymentMethodSvc *serviceMocks.IPaymentMethodService,
			merchantRepo *repositoryMocks.IMerchantRepository,
			orchSvc *serviceMocks.IOrchestratorService)
		input creditcardModel.CardPaymentNotificationRequest
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(
				rmqExtMock *rabbitMqMocks.RabbitMQExt,
				paymentSvc *serviceMocks.IPaymentService,
				mockPaymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchSvc *serviceMocks.IOrchestratorService) {
				// Body Function
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(creditCardDB, nil)
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				orchSvc.On("FindByReference", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(nil, nil)
				paymentSvc.On("PostCreateLedger", mock.Anything, mock.AnythingOfType("*paymentModel.Payment"), constant.PtrPostCreateLedgerRequestMockType()).Return(nil)
				mockPaymentRepo.On("CommitTransaction", mock.Anything).Return(nil)
				rmqExtMock.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmqExtMock.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(merchantData, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
			},
			input: validRequest,
		},
		{
			name: "ERROR:Credit Card Not Found",
			mockSetup: func(
				rmqExtMock *rabbitMqMocks.RabbitMQExt,
				paymentSvc *serviceMocks.IPaymentService,
				mockPaymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(nil, nil).Once()
				mockPaymentRepo.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(paymentMethodDB, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(nil, nil).Once()
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(func() *paymentModel.Payment {
					payment := *paymentDB
					payment.Type = constant.TypeVirtualTerminal
					return &payment
				}(), nil).Once()
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				orchSvc.On("FindByReference", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(nil, nil)
				paymentSvc.On("PostCreateLedger",
					mock.Anything,
					mock.AnythingOfType("*paymentModel.Payment"),
					constant.PtrPostCreateLedgerRequestMockType(),
				).Return(nil)
				mockPaymentRepo.On("CommitTransaction", mock.Anything).Return(nil)
				orchSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).Return(nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
			input:   validRequest,
		},
		{
			name: "ERROR:FindByReferenceID Failure",
			mockSetup: func(_ *rabbitMqMocks.RabbitMQExt, _ *serviceMocks.IPaymentService, mockPaymentRepo *repositoryMocks.IPaymentRepository, _ *serviceMocks.IPaymentMethodService, _ *repositoryMocks.IMerchantRepository, _ *serviceMocks.IOrchestratorService) {
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR:Find merchant by id",
			mockSetup: func(_ *rabbitMqMocks.RabbitMQExt, _ *serviceMocks.IPaymentService, mockPaymentRepo *repositoryMocks.IPaymentRepository, _ *serviceMocks.IPaymentMethodService, merchantRepo *repositoryMocks.IMerchantRepository, _ *serviceMocks.IOrchestratorService) {

				// Body Function
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(creditCardDB, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR:Determine payment fee",
			mockSetup: func(_ *rabbitMqMocks.RabbitMQExt, paymentSvc *serviceMocks.IPaymentService, mockPaymentRepo *repositoryMocks.IPaymentRepository, _ *serviceMocks.IPaymentMethodService, merchantRepo *repositoryMocks.IMerchantRepository, _ *serviceMocks.IOrchestratorService) {

				// Body Function
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(creditCardDB, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(merchantData, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR:Update Failure",
			mockSetup: func(_ *rabbitMqMocks.RabbitMQExt, paymentSvc *serviceMocks.IPaymentService, mockPaymentRepo *repositoryMocks.IPaymentRepository, _ *serviceMocks.IPaymentMethodService, merchantRepo *repositoryMocks.IMerchantRepository, _ *serviceMocks.IOrchestratorService) {

				// Body Function
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(creditCardDB, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(merchantData, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(constant.ErrSomeErrorForUnitTest)
				mockPaymentRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			wantErr: true,
			input:   validRequest,
		},
	}
	buf := new(bytes.Buffer)
	defer buf.Reset()

	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			rmqExtMock := rabbitMqMocks.NewRabbitMQExt(t)
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
			paymentSvc := serviceMocks.NewIPaymentService(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			orchSvc := serviceMocks.NewIOrchestratorService(t)
			ctx := context.Background()

			tc.mockSetup(rmqExtMock, paymentSvc, mockPaymentRepo, paymentMethodSvc, merchantRepo, orchSvc)

			svc := New(conf, log, rmqExtMock, mockPaymentRepo, nil, nil, WithPaymentLedgerService(paymentSvc), WithMerchantRepo(merchantRepo), WithOrchestratorService(orchSvc), WithPaymentMethodService(paymentMethodSvc))
			err := svc.PaymentNotification(ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			orchSvc.AssertExpectations(t)
			paymentSvc.AssertExpectations(t)
			rmqExtMock.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			paymentMethodSvc.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}
