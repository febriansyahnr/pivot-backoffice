package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCapture(t *testing.T) {
	tests := []struct {
		name             string
		paymentID        string
		merchantClaim    *merchant.MerchantAuthTokenClaims
		requestBody      interface{}
		requestHeader    map[string]string
		expectedCode     int
		expectedResponse string
		setup            func(mockService *serviceMocks.IUnifiedPaymentService)
	}{
		{
			name:          "ERROR: Merchant not found",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: nil,
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    50000,
				},
			},
			expectedCode:     http.StatusUnauthorized,
			expectedResponse: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Invalid payment UUID",
			paymentID:     "invalid-uuid",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    50000,
				},
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"payment id is not valid"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Invalid JSON",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:   "invalid json",
			expectedCode:  http.StatusBadRequest,
			expectedResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Amount required when releaseRemainingAmount is false",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"releaseRemainingAmount": false,
			},
			expectedCode:     http.StatusUnprocessableEntity,
			expectedResponse: `{"code":"unprocessable_entity","message":"Unprocessable entity","error":{"type":"API_ERROR","details":[{"field":"","message":"amount is required when releaseRemainingAmount is false"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Amount value must be greater than 0",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    0,
				},
			},
			expectedCode:     http.StatusUnprocessableEntity,
			expectedResponse: `{"code":"unprocessable_entity","message":"Unprocessable entity","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value must be greater than 0"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Amount value negative",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    -100,
				},
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'CaptureRequest.Amount.Value' Error:Field validation for 'Value' failed on the 'min' tag"}],"traceId":""}}`,
		},
		{
			name:          "ERROR: Service returns error",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    50000,
				},
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "550e8400-e29b-41d4-a716-446655440000" &&
						req.MerchantID == "merchant-123" &&
						req.Amount != nil &&
						req.Amount.Currency == "IDR" &&
						req.Amount.Value == 50000
				})).Return(nil, errors.New("service error")).Once()
			},
		},
		{
			name:          "SUCCESS: Capture with amount",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    50000,
				},
			},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","data":{"amount":{"currency":"IDR","value":50000},"createdAt":"0001-01-01T00:00:00Z","id":"550e8400-e29b-41d4-a716-446655440000","paymentSessionClientReferenceId":"","paymentSessionId":"","releaseRemainingAmount":false,"status":"SUCCESS","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "550e8400-e29b-41d4-a716-446655440000" &&
						req.MerchantID == "merchant-123" &&
						req.Amount != nil &&
						req.Amount.Currency == "IDR" &&
						req.Amount.Value == 50000
				})).Return(&unifiedPaymentModel.CaptureResponse{
					ID: "550e8400-e29b-41d4-a716-446655440000",
					Amount: &unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    50000,
					},
					Status: "SUCCESS",
				}, nil).Once()
			},
		},
		{
			name:          "SUCCESS: Capture with releaseRemainingAmount true",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"releaseRemainingAmount": true,
			},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","data":{"amount":null,"createdAt":"0001-01-01T00:00:00Z","id":"550e8400-e29b-41d4-a716-446655440000","paymentSessionClientReferenceId":"","paymentSessionId":"","releaseRemainingAmount":true,"status":"SUCCESS","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "550e8400-e29b-41d4-a716-446655440000" &&
						req.MerchantID == "merchant-123" &&
						req.ReleaseRemainingAmount == true
				})).Return(&unifiedPaymentModel.CaptureResponse{
					ID:                     "550e8400-e29b-41d4-a716-446655440000",
					ReleaseRemainingAmount: true,
					Status:                 "SUCCESS",
				}, nil).Once()
			},
		},
		{
			name:          "SUCCESS: Capture with sub-merchant ID",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    30000,
				},
			},
			requestHeader:    map[string]string{constant.HeaderXSubMerchantID: "sub-merchant-789"},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","data":{"amount":{"currency":"IDR","value":30000},"createdAt":"0001-01-01T00:00:00Z","id":"550e8400-e29b-41d4-a716-446655440000","paymentSessionClientReferenceId":"","paymentSessionId":"","releaseRemainingAmount":false,"status":"SUCCESS","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "550e8400-e29b-41d4-a716-446655440000" &&
						req.MerchantID == "sub-merchant-789" &&
						req.Amount != nil &&
						req.Amount.Value == 30000
				})).Return(&unifiedPaymentModel.CaptureResponse{
					ID: "550e8400-e29b-41d4-a716-446655440000",
					Amount: &unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    30000,
					},
					Status: "SUCCESS",
				}, nil).Once()
			},
		},
		{
			name:          "SUCCESS: Capture with releaseRemainingAmount true and amount provided",
			paymentID:     "550e8400-e29b-41d4-a716-446655440000",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody: map[string]interface{}{
				"releaseRemainingAmount": true,
				"amount": map[string]interface{}{
					"currency": "IDR",
					"value":    75000,
				},
			},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","data":{"amount":{"currency":"IDR","value":75000},"createdAt":"0001-01-01T00:00:00Z","id":"550e8400-e29b-41d4-a716-446655440000","paymentSessionClientReferenceId":"","paymentSessionId":"","releaseRemainingAmount":true,"status":"SUCCESS","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "550e8400-e29b-41d4-a716-446655440000" &&
						req.MerchantID == "merchant-123" &&
						req.ReleaseRemainingAmount == true &&
						req.Amount != nil &&
						req.Amount.Value == 75000
				})).Return(&unifiedPaymentModel.CaptureResponse{
					ID: "550e8400-e29b-41d4-a716-446655440000",
					Amount: &unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    75000,
					},
					ReleaseRemainingAmount: true,
					Status:                 "SUCCESS",
				}, nil).Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIUnifiedPaymentService(t)
			if tc.setup != nil {
				tc.setup(mockService)
			}

			cfg := &config.Config{
				Environment: "test",
			}
			controller := New(cfg, nil, WithUnifiedPaymentService(mockService))

			var reqBody []byte
			var err error
			if tc.requestBody != nil {
				if str, ok := tc.requestBody.(string); ok {
					// Handle invalid JSON test case
					reqBody = []byte(str)
				} else {
					reqBody, err = json.Marshal(tc.requestBody)
					if err != nil {
						t.Fatalf("failed to marshal request body: %v", err)
					}
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/payments/"+tc.paymentID+"/capture", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, merchantPlatformWhitelistedOldResponseFormat)
			if tc.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, tc.merchantClaim)
			}

			for key, value := range tc.requestHeader {
				req.Header.Set(key, value)
			}

			if tc.paymentID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("uuid", tc.paymentID)
				ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
			}
			req = req.WithContext(ctx)

			controller.Capture(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedResponse, rec.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
