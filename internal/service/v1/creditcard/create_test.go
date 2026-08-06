package creditcard_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	paymentMethodConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/creditcard"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreatePayment(t *testing.T) {
	now := time.Now()
	expiredAt := now.Add(constant.CreditCardPaymentExpired)
	merchantID := uuid.New()
	paymentMethodID := uuid.New()
	referenceID := "reference-id"

	validRequest := creditcardModel.CreateCardPaymentRequest{
		MerchantID:           merchantID,
		BankMerchantID:       "bank-merchant-id",
		ReferenceID:          referenceID,
		Amount:               decimal.NewFromFloat(100.0),
		Currency:             "IDR",
		AuthenticationMethod: "CHALLENGE",
	}
	validRequestWithSplitRouteConfig := creditcardModel.CreateCardPaymentRequest{
		MerchantID:           merchantID,
		BankMerchantID:       "bank-merchant-id",
		ReferenceID:          referenceID,
		Amount:               decimal.NewFromFloat(100.0),
		Currency:             "IDR",
		AuthenticationMethod: "CHALLENGE",
		SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
			{
				MerchantId:  "73c39ae0-a2a9-4edf-84b5-3405c6832f93",
				Type:        "FIXED", // NOSONAR
				Currency:    "IDR",   // NOSONAR
				FixedAmount: 1_000,   // NOSONAR
			},
		},
	}
	config := &config.Config{
		CreditcardConfig: config.CreditcardConfig{
			WebviewURL: "https://example.com",
		},
	}

	paymentCCDB := &paymentModel.Payment{
		ReferenceID: &referenceID,
		Status:      constant.CreditCardStatusWaitingForPayment,
		PaymentURL:  "https://example.com/pay/existing",
		ExpiredAt:   &expiredAt,
		CreatedAt:   now,
	}

	paymentCCProcessedDB := &paymentModel.Payment{
		ReferenceID: &referenceID,
		Status:      constant.CreditCardStatusPAID,
		PaymentURL:  "https://example.com/pay/existing",
		ExpiredAt:   &expiredAt,
		CreatedAt:   now,
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
	ctxValue := context.WithValue(context.Background(), constant.CtxTraceIdKey, "af1c9289-7536-42f2-bd80-f6980e5c9d99")

	merchantData := &merchantModel.Merchant{}

	testCases := []struct {
		name      string
		input     creditcardModel.CreateCardPaymentRequest
		ctx       context.Context
		wantErr   bool
		mockSetup func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService, merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService)
	}{
		{
			name:    "SUCCESS: New Payment",
			wantErr: false,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						Valid:  true,
						String: "f78ab91e-d1c5-47f9-b6bd-6f7677ce38f4",
					},
				}, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(paymentMethodDB, nil)
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				orchSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).Return(nil)
				mockPaymentRepo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			input: validRequest,
			ctx:   ctxValue,
		},
		{
			name:    "SUCCESS: New Payment With Split Route Configuration",
			wantErr: false,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						Valid:  true,
						String: "f78ab91e-d1c5-47f9-b6bd-6f7677ce38f4",
					},
				}, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(paymentMethodDB, nil)
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				orchSvc.On("PostAccountTransaction", mock.Anything, constant.PtrCreateAccTransactionReqMockType()).Return(nil)
				mockPaymentRepo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			input: validRequestWithSplitRouteConfig,
			ctx:   ctxValue,
		},
		{
			name: "ERROR: Merchant Error",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Merchant Not Found",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Payment Already Exists",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(paymentCCDB, nil)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Duplicate Payment",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(paymentCCProcessedDB, nil)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Find Payment Failure",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Get Payment Method Creditcard",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Payment Method Creditcard Not Found",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantErr: true,
			input:   validRequest,
		},
		{
			name: "ERROR: Split route config do not apply to facilitator payment methods ",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					ChannelType: constant.PaymentMethodChannelTypeDirect,
				}, nil)
			},
			wantErr: true,
			input:   validRequestWithSplitRouteConfig,
		},
		{
			name: "ERROR: Insert Failure",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockPaymentMethodSvc *serviceMocks.IPaymentMethodService,
				merchantRepo *repositoryMocks.IMerchantRepository, orchSvc *serviceMocks.IOrchestratorService) {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchantData, nil)
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
				mockPaymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Return(paymentMethodDB, nil)
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				mockPaymentRepo.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(constant.ErrSomeErrorForUnitTest)
				mockPaymentRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			wantErr: true,
			input:   validRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockPaymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

			if tc.ctx == nil {
				tc.ctx = context.WithValue(context.Background(), constant.CtxParentMerchantId, uuid.NewString())
			}
			tc.mockSetup(mockPaymentRepo, mockPaymentMethodSvc, merchantRepo, orchestratorSvc)

			svc := New(config, mockLogger, nil, mockPaymentRepo, nil, nil, WithMerchantRepo(merchantRepo), WithOrchestratorService(orchestratorSvc), WithPaymentMethodService(mockPaymentMethodSvc))
			response, err := svc.CreatePayment(tc.ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				if response != nil {
					assert.Equal(t, tc.input.ReferenceID, response.ReferenceID)
				}
			}

			mockPaymentRepo.AssertExpectations(t)
			mockPaymentMethodSvc.AssertExpectations(t)
		})
	}
}
