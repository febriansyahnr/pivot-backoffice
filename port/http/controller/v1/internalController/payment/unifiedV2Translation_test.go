package internalPaymentController

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
)

func TestInternalPaymentController_createPaymentViaUnifiedV2(t *testing.T) {
	mockUnifiedPaymentSvc := &serviceMocks.IUnifiedPaymentService{}
	mockLogger := &loggerMocks.ILogger{}

	controller := &InternalPaymentController{
		unifiedPaymentSvc: mockUnifiedPaymentSvc,
		logger:            mockLogger,
	}

	// Setup tracer for testing
	otelTracer = otel.Tracer("test-tracer")

	tests := []struct {
		name        string
		setupMocks  func()
		merchantID  string
		snapRequest paymentModel.PaymentRequest
		wantErr     bool
		expectedErr string
	}{
		{
			name: "success - create payment via unified v2",
			setupMocks: func() {
				mockUnifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						ID:                "test-payment-id",
						ClientReferenceID: "REF123",
						Status:            constant.StatusActive,
						Amount: unifiedPaymentModel.Amount{
							Value:    10000.0,
							Currency: "IDR",
						},
					}, nil).Once()
			},
			merchantID: "merchant-123",
			snapRequest: paymentModel.PaymentRequest{
				UUID:          "payment-uuid",
				ReferenceID:   "REF123",
				PaymentMethod: "QRIS",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(10000),
					Currency: "IDR",
				},
				CreatedBy: "test-user",
			},
			wantErr: false,
		},
		{
			name: "error - unified payment service is nil",
			setupMocks: func() {
				// No mocks needed
			},
			merchantID: "merchant-123",
			snapRequest: paymentModel.PaymentRequest{
				UUID:        "payment-uuid",
				ReferenceID: "REF123",
			},
			wantErr:     true,
			expectedErr: "unified Payment V2 service unavailable",
		},
		{
			name: "error - create session fails",
			setupMocks: func() {
				mockUnifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).
					Return(nil, errors.New("create session failed")).Once()
			},
			merchantID: "merchant-123",
			snapRequest: paymentModel.PaymentRequest{
				UUID:          "payment-uuid",
				ReferenceID:   "REF123",
				PaymentMethod: "QRIS",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(10000),
					Currency: "IDR",
				},
			},
			wantErr:     true,
			expectedErr: "create session failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUnifiedPaymentSvc.ExpectedCalls = nil
			mockLogger.ExpectedCalls = nil

			// Setup controller for this test
			testController := controller
			if tt.name == "error - unified payment service is nil" {
				testController = &InternalPaymentController{
					unifiedPaymentSvc: nil,
					logger:            mockLogger,
				}
				// Mock the logger error call for this test
				mockLogger.On("Error", mock.Anything, "Unified Payment V2 service is not initialized").Once()
			}

			// Mock logger for other error cases
			if tt.name == "error - create session fails" {
				mockLogger.On("Error", mock.Anything, "Failed to create payment via Unified Payment V2", mock.Anything).Once()
			}

			tt.setupMocks()

			result, err := testController.createPaymentViaUnifiedV2(context.Background(), tt.merchantID, tt.snapRequest)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test-payment-id", result.UUID)
				assert.Equal(t, "REF123", result.ReferenceID)
				assert.Equal(t, constant.StatusPending, result.Status) // ACTIVE is mapped to PENDING
				assert.Equal(t, constant.UnifiedPaymentTypeMultiple, result.PaymentType)
				assert.False(t, result.IsUnifiedPayment) // Keep false for SNAP API compatibility
			}

			mockUnifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestInternalPaymentController_translateSnapToUnifiedRequest(t *testing.T) {
	controller := &InternalPaymentController{}
	merchantID := "merchant-123"

	tests := []struct {
		name        string
		snapRequest paymentModel.PaymentRequest
		validate    func(t *testing.T, result *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest)
	}{
		{
			name: "basic translation - QRIS payment",
			snapRequest: paymentModel.PaymentRequest{
				UUID:          "payment-uuid",
				ReferenceID:   "REF123",
				PaymentMethod: "QRIS",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(50000),
					Currency: "IDR",
				},
				CreatedBy: "test-user",
				Qris: &paymentModel.PaymentMetadataQris{
					QrType: "STATIC",
				},
				Customer: paymentModel.PaymentRequestCustomer{
					CustomerID: "customer-123",
					Email:      "test@example.com",
				},
				ClientRedirectUrl: paymentModel.UnifiedPaymentRedirectUrl{
					SuccessUrl: "https://success.com",
					FailedUrl:  "https://failed.com",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				assert.Equal(t, "REF123", result.ClientReferenceID)
				assert.Equal(t, constant.UnifiedPaymentTypeMultiple, result.PaymentType)
				assert.Equal(t, 50000.0, result.Amount.Value)
				assert.Equal(t, "IDR", result.Amount.Currency)
				assert.Equal(t, constant.UnifiedPaymentModeRedirect, result.Mode)
				assert.True(t, result.AutoConfirm)
				assert.Equal(t, "payment-uuid", result.PaymentID)
				assert.Equal(t, "merchant-123", result.MerchantID)
				assert.Equal(t, "test-user", result.CreatedBy)
				assert.True(t, result.IsMigratingFromV1)

				assert.NotNil(t, result.PaymentMethod)
				assert.Equal(t, "QR", result.PaymentMethod.Type)
				assert.NotNil(t, result.PaymentMethodOptions.QR)

				assert.Equal(t, "customer-123", result.CustomerID)
				assert.NotNil(t, result.CustomerInformation)
				assert.Equal(t, "test@example.com", result.CustomerInformation.Email)

				assert.Equal(t, "https://success.com", result.RedirectUrl.SuccessReturnUrl)
				assert.Equal(t, "https://failed.com", result.RedirectUrl.FailureReturnUrl)

				// Note: Metadata is not set by translateSnapToUnifiedRequest function
				assert.Nil(t, result.Metadata)
			},
		},
		{
			name: "virtual account payment",
			snapRequest: paymentModel.PaymentRequest{
				UUID:          "payment-uuid",
				ReferenceID:   "REF456",
				PaymentMethod: "VIRTUAL_ACCOUNT",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(25000),
					Currency: "IDR",
				},
				VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
					Issuer: "BCA",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				assert.Equal(t, "REF456", result.ClientReferenceID)
				assert.Equal(t, 25000.0, result.Amount.Value)
				assert.NotNil(t, result.PaymentMethod)
				assert.Equal(t, "VIRTUAL_ACCOUNT", result.PaymentMethod.Type)
				assert.NotNil(t, result.PaymentMethodOptions.VirtualAccount)
				assert.Equal(t, "BCA", result.PaymentMethodOptions.VirtualAccount.Channel)

				// Note: Metadata is not set by translateSnapToUnifiedRequest function
				assert.Nil(t, result.Metadata)
			},
		},
		{
			name: "minimal request - no payment method",
			snapRequest: paymentModel.PaymentRequest{
				UUID:        "payment-uuid",
				ReferenceID: "REF789",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(10000),
					Currency: "IDR",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				assert.Equal(t, "REF789", result.ClientReferenceID)
				assert.Equal(t, 10000.0, result.Amount.Value)
				assert.Nil(t, result.PaymentMethod)
				assert.Empty(t, result.CustomerID)
				assert.Nil(t, result.CustomerInformation)
			},
		},
		{
			name: "with split routing configurations",
			snapRequest: paymentModel.PaymentRequest{
				UUID:        "payment-uuid",
				ReferenceID: "REF999",
				TotalAmount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(100000),
					Currency: "IDR",
				},
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:  "merchant-123",
						Type:        "FIXED",
						Currency:    "IDR",
						FixedAmount: 10000,
						Remarks:     "Test split",
					},
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				assert.Equal(t, "REF999", result.ClientReferenceID)
				assert.NotNil(t, result.SplitRoutingConfigurations)
				assert.Len(t, *result.SplitRoutingConfigurations, 1)
				assert.Equal(t, "merchant-123", (*result.SplitRoutingConfigurations)[0].MerchantId)
				assert.Equal(t, float64(10000), (*result.SplitRoutingConfigurations)[0].FixedAmount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.translateSnapToUnifiedRequest(tt.snapRequest, merchantID)

			assert.NotNil(t, result)
			tt.validate(t, result)
		})
	}
}

func TestInternalPaymentController_translateUnifiedToSnapResponse(t *testing.T) {
	controller := &InternalPaymentController{}

	tests := []struct {
		name        string
		unifiedResp *unifiedPaymentModel.UnifiedPaymentSessionResponse
		snapRequest paymentModel.PaymentRequest
		validate    func(t *testing.T, result *paymentModel.PaymentResponse)
	}{
		{
			name: "complete response with virtual account",
			unifiedResp: &unifiedPaymentModel.UnifiedPaymentSessionResponse{
				ID:                "unified-payment-123",
				ClientReferenceID: "REF123",
				Status:            constant.StatusActive,
				Amount: unifiedPaymentModel.Amount{
					Value:    50000.0,
					Currency: "IDR",
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					VAPaymentMethodDetail: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
						VirtualAccountNumber: "1234567890123456",
						VirtualAccountName:   "Test Account",
					},
				},
			},
			snapRequest: paymentModel.PaymentRequest{
				PaymentMethod: "VIRTUAL_ACCOUNT",
				VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
					Issuer: "BCA",
				},
				Customer: paymentModel.PaymentRequestCustomer{
					CustomerID: "customer-123",
					Name:       "John Doe",
					Email:      "john@example.com",
					Phone:      "081234567890",
				},
				PaymentItems: &[]paymentModel.PaymentItemRequest{
					{
						Name:        "Product 1",
						Description: "Test product",
						Qty:         2,
						Amount: paymentModel.Amount{
							Value:    decimal.NewFromFloat(25000),
							Currency: "IDR",
						},
					},
				},
			},
			validate: func(t *testing.T, result *paymentModel.PaymentResponse) {
				assert.Equal(t, "unified-payment-123", result.UUID)
				assert.Equal(t, "REF123", result.ReferenceID)
				assert.Equal(t, constant.StatusPending, result.Status) // ACTIVE is mapped to PENDING
				assert.Equal(t, "VIRTUAL_ACCOUNT", result.PaymentMethod)
				assert.Equal(t, constant.UnifiedPaymentTypeMultiple, result.PaymentType)
				assert.False(t, result.IsUnifiedPayment) // Keep false for SNAP API compatibility

				assert.NotNil(t, result.TotalAmount)
				assert.Equal(t, "50000", result.TotalAmount.Value.String())
				assert.Equal(t, "IDR", result.TotalAmount.Currency)

				assert.NotNil(t, result.Customer)
				assert.Equal(t, "customer-123", result.Customer.CustomerID)
				assert.Equal(t, "John Doe", result.Customer.Name)
				assert.Equal(t, "john@example.com", result.Customer.Email)

				assert.NotNil(t, result.VirtualAccount)
				assert.Equal(t, "1234567890123456", result.VirtualAccount.VirtualAccountNumber)
				assert.Equal(t, "Test Account", result.VirtualAccount.VirtualAccountName)

				assert.NotNil(t, result.PaymentItems)
				assert.Len(t, *result.PaymentItems, 1)
				assert.Equal(t, "Product 1", (*result.PaymentItems)[0].Name)
				assert.Equal(t, "Test product", (*result.PaymentItems)[0].Description)
				assert.Equal(t, 2, (*result.PaymentItems)[0].Qty)
			},
		},
		{
			name: "QRIS payment response",
			unifiedResp: &unifiedPaymentModel.UnifiedPaymentSessionResponse{
				ID:                "unified-payment-456",
				ClientReferenceID: "REF456",
				Status:            "REQUIRE_ACTION",
				Amount: unifiedPaymentModel.Amount{
					Value:    25000.0,
					Currency: "IDR",
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					QrPaymentMethodDetail: &unifiedPaymentModel.ChargePaymentMethodDetailQr{
						QrContent: "00020101021126...QR_CONTENT_HERE",
					},
				},
			},
			snapRequest: paymentModel.PaymentRequest{
				PaymentMethod: "QRIS",
				Qris: &paymentModel.PaymentMetadataQris{
					QrType: "STATIC",
				},
				Customer: paymentModel.PaymentRequestCustomer{
					CustomerID: "customer-456",
					Email:      "test@example.com",
				},
			},
			validate: func(t *testing.T, result *paymentModel.PaymentResponse) {
				assert.Equal(t, "unified-payment-456", result.UUID)
				assert.Equal(t, "REF456", result.ReferenceID)
				assert.Equal(t, "REQUIRE_ACTION", result.Status)
				assert.Equal(t, "QRIS", result.PaymentMethod)

				assert.NotNil(t, result.Qris)
				assert.Equal(t, "00020101021126...QR_CONTENT_HERE", result.Qris.QrContent)

				assert.Equal(t, "25000", result.TotalAmount.Value.String())
			},
		},
		{
			name: "minimal response - no payment method details",
			unifiedResp: &unifiedPaymentModel.UnifiedPaymentSessionResponse{
				ID:                "unified-payment-789",
				ClientReferenceID: "REF789",
				Status:            constant.StatusPending,
				Amount: unifiedPaymentModel.Amount{
					Value:    10000.0,
					Currency: "IDR",
				},
				PaymentMethod: nil,
			},
			snapRequest: paymentModel.PaymentRequest{
				PaymentMethod: "CREDIT_CARD",
				Customer: paymentModel.PaymentRequestCustomer{
					CustomerID: "customer-789",
				},
			},
			validate: func(t *testing.T, result *paymentModel.PaymentResponse) {
				assert.Equal(t, "unified-payment-789", result.UUID)
				assert.Equal(t, "REF789", result.ReferenceID)
				assert.Equal(t, constant.StatusPending, result.Status)
				assert.Equal(t, "CREDIT_CARD", result.PaymentMethod)

				assert.Nil(t, result.VirtualAccount)
				assert.Nil(t, result.Qris)
				assert.Equal(t, "10000", result.TotalAmount.Value.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.translateUnifiedToSnapResponse(tt.unifiedResp, tt.snapRequest)

			assert.NotNil(t, result)
			tt.validate(t, result)
		})
	}
}
