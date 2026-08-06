package v1InternalUnifiedPaymentController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	m.Run()
}

func TestCreate(t *testing.T) {
	merchantID := uuid.NewString()
	futureTime := time.Now().Add(24 * time.Hour)

	validPayload := paymentModel.CreateUnifiedPaymentRequest{
		MerchantID:        merchantID,
		ClientReferenceID: "ref-123",
		PaymentMethod:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Amount: paymentModel.Amount{
			Currency: "IDR",
			Value:    decimal.NewFromInt(100000),
		},
		ExpiryAt: futureTime,
		Customer: paymentModel.PaymentRequestCustomer{
			Name:  "John Doe",
			Email: "john@example.com",
		},
		PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
			VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
				Issuer: "BCA",
			},
		},
	}

	tests := []struct {
		name         string
		setupMock    func(*serviceMocks.IPaymentService, *serviceMocks.IUnifiedPaymentService, *serviceMocks.ICustomerService)
		setupRequest func() *http.Request
		setupContext func(context.Context) context.Context
		expectedCode int
		checkBody    bool
	}{
		{
			name: "SUCCESS: Missing merchant auth",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Invalid JSON",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBufferString("invalid json"))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation error",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
			},
			setupRequest: func() *http.Request {
				invalidPayload := paymentModel.CreateUnifiedPaymentRequest{
					MerchantID:    merchantID,
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					// Missing required fields
				}
				body, _ := json.Marshal(invalidPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Expiry in past",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
			},
			setupRequest: func() *http.Request {
				pastPayload := validPayload
				pastPayload.ExpiryAt = time.Now().Add(-1 * time.Hour)
				body, _ := json.Marshal(pastPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "SUCCESS: V1 flow (feature flag disabled)",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				ps.On("CreateUnifiedPayment", constant.ValueCtxMockType(), mock.Anything).
					Return(&paymentModel.CreateUnifiedPaymentResponse{}, nil).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusOK,
			checkBody:    true,
		},
		{
			name: "ERROR: V1 flow service error",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				ps.On("CreateUnifiedPayment", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, errors.New("service error")).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "SUCCESS: With sub-merchant ID header",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				ps.On("CreateUnifiedPayment", constant.ValueCtxMockType(), mock.Anything).
					Return(&paymentModel.CreateUnifiedPaymentResponse{}, nil).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				req := httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
				req.Header.Set(constant.HeaderXSubMerchantID, "sub-merchant-123")
				return req
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "SUCCESS: V2 flow with feature flag enabled",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				// Mock customer creation since V2 flow creates customer from info
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-123"}, nil).Once()
				ups.On("CreateSession", constant.ValueCtxMockType(), mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						ID:                "session-123",
						ClientReferenceID: "ref-123",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						PaymentMethod: &unifiedPaymentModel.PaymentMethod{
							Type: constant.UnifiedPaymentMethodVA,
						},
						PaymentUrl: "https://payment.url",
					}, nil).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: "test-v2-merchant-123",
				})
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "ERROR: V2 flow - returns error when validation fails",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				// Mock customer creation to fail
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, errors.New("customer creation failed")).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: "test-v2-merchant-123",
				})
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "ERROR: V2 flow with service error",
			setupMock: func(ps *serviceMocks.IPaymentService, ups *serviceMocks.IUnifiedPaymentService, cs *serviceMocks.ICustomerService) {
				// Mock customer creation since V2 flow creates customer from info
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-456"}, nil).Once()
				ups.On("CreateSession", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, errors.New("service error")).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPost, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: "test-v2-merchant-123",
				})
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentSvc := serviceMocks.NewIPaymentService(t)
			mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockCustomerSvc := serviceMocks.NewICustomerService(t)
			mockLogger := loggerMocks.NewILogger(t)

			tc.setupMock(mockPaymentSvc, mockUnifiedPaymentSvc, mockCustomerSvc)

			controller := &paymentController{
				config:            &config.Config{Environment: "test"},
				validate:          validatorExt.New(),
				paymentSvc:        mockPaymentSvc,
				unifiedPaymentSvc: mockUnifiedPaymentSvc,
				customerSvc:       mockCustomerSvc,
				logger:            mockLogger,
			}

			req := tc.setupRequest()
			ctx := tc.setupContext(req.Context())
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.Create)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "Response code mismatch for test: %s", tc.name)
		})
	}
}

func TestValidatePayload(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		payload *paymentModel.CreateUnifiedPaymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "SUCCESS: Valid VA payload",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt: futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "BCA",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Expiry in the past",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt: pastTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "BCA",
					},
				},
			},
			wantErr: true,
			errMsg:  "expiry is not allowed to be less than current time",
		},
		{
			name: "ERROR: VA without VA options",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt:             futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{},
			},
			wantErr: true,
			errMsg:  "payment method option virtual account is required",
		},
		{
			name: "ERROR: Credit card without card options",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt:             futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{},
			},
			wantErr: true,
			errMsg:  "payment method option card is required",
		},
		{
			name: "ERROR: Split routing currency mismatch",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt: futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "BCA",
					},
				},
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						Currency:    "USD",
						FixedAmount: 50000,
						Type:        constant.SplitRoutingPaymentTypeFixed,
					},
				},
			},
			wantErr: true,
			errMsg:  "currency is not match",
		},
		{
			name: "ERROR: Split routing exceeds amount",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt: futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "BCA",
					},
				},
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						Currency:    "IDR",
						FixedAmount: 150000,
						Type:        constant.SplitRoutingPaymentTypeFixed,
					},
				},
			},
			wantErr: true,
			errMsg:  "total split and routing amount must be not greater than payment amount",
		},
		{
			name: "SUCCESS: Split routing with percentage",
			payload: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Amount: paymentModel.Amount{
					Currency: "IDR",
					Value:    decimal.NewFromInt(100000),
				},
				ExpiryAt: futureTime,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "BCA",
					},
				},
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						Currency:         "IDR",
						PercentageAmount: 50,
						Type:             constant.SplitRoutingPaymentTypePercentage,
					},
				},
			},
			wantErr: false,
		},
	}

	controller := &paymentController{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := controller.validatePayload(tc.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCustomerPayload(t *testing.T) {
	merchantID := "merchant-123"
	customerID := "customer-123"

	validCustomerInfo := &unifiedPaymentModel.CustomerInformation{
		GivenName: "John",
		Surname:   util.ValueToPtr("Doe"),
		Email:     "john@example.com",
		RefundPreference: &unifiedPaymentModel.UnifiedPaymentRefundPreference{
			Method: "bank_transfer",
			TransferDestination: &unifiedPaymentModel.RefundTransferDestination{
				ChannelCode: "PERMATA",
				ChannelInformation: &unifiedPaymentModel.RefundChannelInformation{
					AccountNumber: "1234567890",
					AccountName:   "John Doe",
				},
			},
		},
	}

	validCustomerInfoWithPhone := *validCustomerInfo
	validCustomerInfoWithPhone.PhoneNumber = &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
		CountryCode: "+62",
		Number:      "8123456789",
	}

	tests := []struct {
		name      string
		payload   *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
		setupMock func(*serviceMocks.ICustomerService)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "SUCCESS: No customer info",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: merchantID,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {},
			wantErr:   false,
		},
		{
			name: "ERROR: Both CustomerID and CustomerInformation",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          merchantID,
				CustomerID:          customerID,
				CustomerInformation: validCustomerInfo,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {},
			wantErr:   true,
			errMsg:    "customer information conflict",
		},
		{
			name: "SUCCESS: Create customer with info",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          merchantID,
				CustomerInformation: validCustomerInfo,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-123"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Create customer with phone",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          merchantID,
				CustomerInformation: &validCustomerInfoWithPhone,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-456"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: Create customer fails",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          merchantID,
				CustomerInformation: validCustomerInfo,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, errors.New("failed to create customer")).Once()
			},
			wantErr: true,
			errMsg:  "failed to create customer",
		},
		{
			name: "SUCCESS: Existing customer not blocked",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: merchantID,
				CustomerID: customerID,
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("GetCustomerById", constant.ValueCtxMockType(), customerID, merchantID).
					Return(&customerModel.GeneralCustomerResponse{
						UUID:      customerID,
						IsBlocked: false,
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: Customer not found",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: merchantID,
				CustomerID: "non-existent",
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("GetCustomerById", constant.ValueCtxMockType(), "non-existent", merchantID).
					Return(nil, nil).Once()
			},
			wantErr: true,
			errMsg:  "customer not found",
		},
		{
			name: "ERROR: GetCustomerById error",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: merchantID,
				CustomerID: "error-customer",
			},
			setupMock: func(cs *serviceMocks.ICustomerService) {
				cs.On("GetCustomerById", constant.ValueCtxMockType(), "error-customer", merchantID).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := serviceMocks.NewICustomerService(t)
			mockLogger := loggerMocks.NewILogger(t)

			tc.setupMock(mockCustomerSvc)

			controller := &paymentController{
				customerSvc: mockCustomerSvc,
				logger:      mockLogger,
			}

			ctx := context.Background()
			err := controller.ValidateCustomerPayload(ctx, tc.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}

	t.Run("SUCCESS: Create controller without options", func(t *testing.T) {
		controller := New(cfg, nil)
		assert.NotNil(t, controller)

		// Assert that the controller has the expected fields
		pc, ok := controller.(*paymentController)
		assert.True(t, ok)
		assert.NotNil(t, pc.config)
		assert.NotNil(t, pc.validate)
		assert.Equal(t, cfg, pc.config)
	})

	t.Run("SUCCESS: Create controller with all options", func(t *testing.T) {
		mockPaymentSvc := serviceMocks.NewIPaymentService(t)
		mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
		mockCustomerSvc := serviceMocks.NewICustomerService(t)
		mockLogger := loggerMocks.NewILogger(t)

		controller := New(
			cfg,
			nil,
			WithLogger(mockLogger),
			WithPaymentService(mockPaymentSvc),
			WithUnifiedPaymentService(mockUnifiedPaymentSvc),
			WithCustomerService(mockCustomerSvc),
		)

		assert.NotNil(t, controller)

		// Assert that all dependencies are set correctly
		pc, ok := controller.(*paymentController)
		assert.True(t, ok)
		assert.Equal(t, mockLogger, pc.logger)
		assert.Equal(t, mockPaymentSvc, pc.paymentSvc)
		assert.Equal(t, mockUnifiedPaymentSvc, pc.unifiedPaymentSvc)
		assert.Equal(t, mockCustomerSvc, pc.customerSvc)
	})
}

func TestWithLogger(t *testing.T) {
	mockLogger := loggerMocks.NewILogger(t)
	pc := &paymentController{}

	optionFunc := WithLogger(mockLogger)
	optionFunc(pc)

	assert.Equal(t, mockLogger, pc.logger)
}

func TestWithPaymentService(t *testing.T) {
	mockPaymentSvc := serviceMocks.NewIPaymentService(t)
	pc := &paymentController{}

	optionFunc := WithPaymentService(mockPaymentSvc)
	optionFunc(pc)

	assert.Equal(t, mockPaymentSvc, pc.paymentSvc)
}

func TestWithUnifiedPaymentService(t *testing.T) {
	mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	pc := &paymentController{}

	optionFunc := WithUnifiedPaymentService(mockUnifiedPaymentSvc)
	optionFunc(pc)

	assert.Equal(t, mockUnifiedPaymentSvc, pc.unifiedPaymentSvc)
}

func TestWithCustomerService(t *testing.T) {
	mockCustomerSvc := serviceMocks.NewICustomerService(t)
	pc := &paymentController{}

	optionFunc := WithCustomerService(mockCustomerSvc)
	optionFunc(pc)

	assert.Equal(t, mockCustomerSvc, pc.customerSvc)
}
