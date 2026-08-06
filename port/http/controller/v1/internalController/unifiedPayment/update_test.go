package v1InternalUnifiedPaymentController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	merchantID := "merchant-123"
	futureTime := time.Now().Add(24 * time.Hour)

	validPayload := paymentModel.UpdateUnifiedPaymentRequest{
		MerchantID:        merchantID,
		ClientReferenceID: "ref-123",
		PaymentMethod:     "VIRTUAL_ACCOUNT",
		ExpiryAt:          &futureTime,
		Customer: paymentModel.PaymentRequestCustomer{
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}

	tests := []struct {
		name         string
		setupMock    func(*serviceMocks.IPaymentService)
		setupRequest func() *http.Request
		setupContext func(context.Context) context.Context
		expectedCode int
		checkBody    func(*testing.T, map[string]interface{})
	}{
		{
			name: "ERROR: Missing merchant auth",
			setupMock: func(ps *serviceMocks.IPaymentService) {
				// No mock setup needed
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return ctx // No merchant auth in context
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Invalid JSON",
			setupMock: func(ps *serviceMocks.IPaymentService) {
				// No mock setup needed
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBufferString("invalid json"))
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
			setupMock: func(ps *serviceMocks.IPaymentService) {
				// No mock setup needed
			},
			setupRequest: func() *http.Request {
				invalidPayload := paymentModel.UpdateUnifiedPaymentRequest{
					MerchantID: merchantID,
					// Missing required fields like UUID
				}
				body, _ := json.Marshal(invalidPayload)
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "SUCCESS: V2 migration enabled - deprecated endpoint",
			setupMock: func(ps *serviceMocks.IPaymentService) {
				// No mock setup needed - endpoint returns before calling service
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: "test-v2-merchant-123", // This merchant has V2 migration enabled
				})
			},
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				data, ok := body["data"].(map[string]interface{})
				assert.True(t, ok, "Response should have data field")
				assert.Equal(t, true, data["deprecated"])
				assert.Equal(t, "/v2/payments", data["alternative"])
				assert.Contains(t, data["message"], "deprecated")
			},
		},
		{
			name: "ERROR: Service error",
			setupMock: func(ps *serviceMocks.IPaymentService) {
				ps.On("UpdateUnifiedPayment", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, errors.New("service error")).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "SUCCESS: Update payment",
			setupMock: func(ps *serviceMocks.IPaymentService) {
				futureTime := time.Now().Add(24 * time.Hour)
				ps.On("UpdateUnifiedPayment", constant.ValueCtxMockType(), mock.Anything).
					Return(&paymentModel.UpdateUnifiedPaymentResponse{
						ID:                "payment-id-123",
						ClientReferenceID: "ref-123",
						PaymentMethod:     "VIRTUAL_ACCOUNT",
						ExpiryAt:          futureTime,
						Amount: paymentModel.Amount{
							Currency: "IDR",
							Value:    decimal.NewFromInt(100000),
						},
					}, nil).Once()
			},
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(validPayload)
				return httptest.NewRequest(http.MethodPut, "/internal/v1/payments", bytes.NewBuffer(body))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: merchantID,
				})
			},
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				data, ok := body["data"].(map[string]interface{})
				assert.True(t, ok, "Response should have data field")
				assert.Equal(t, "payment-id-123", data["id"])
				assert.Equal(t, "ref-123", data["clientReferenceId"])
				assert.Equal(t, "VIRTUAL_ACCOUNT", data["paymentMethod"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentSvc := serviceMocks.NewIPaymentService(t)
			mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockCustomerSvc := serviceMocks.NewICustomerService(t)
			mockLogger := loggerMocks.NewILogger(t)

			tc.setupMock(mockPaymentSvc)

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

			handler := http.HandlerFunc(controller.Update)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "Response code mismatch for test: %s", tc.name)

			if tc.checkBody != nil && rr.Code == http.StatusOK {
				var responseBody map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err, "Should be able to parse response body")
				tc.checkBody(t, responseBody)
			}
		})
	}
}
