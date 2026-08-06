package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancel(t *testing.T) {
	tests := []struct {
		name               string
		paymentID          string
		cancellationReason string
		merchantClaim      *merchant.MerchantAuthTokenClaims
		requestBody        map[string]interface{}
		requestHeader      map[string]string
		expectedCode       int
		expectedResponse   string
		wantContains       string
		setup              func(mockService *serviceMocks.IUnifiedPaymentService)
	}{
		{
			name:               "ERROR: Merchant not found",
			paymentID:          "payment-456",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      nil,
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:       http.StatusUnauthorized,
			expectedResponse:   wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name:               "ERROR: Missing payment ID",
			paymentID:          "",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:       http.StatusBadRequest,
			expectedResponse:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid id"}],"traceId":""}}`,
		},
		{
			name:               "ERROR: Invalid JSON",
			paymentID:          "payment-456",
			cancellationReason: "",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        nil, // Will send invalid JSON directly
			expectedCode:       http.StatusBadRequest,
			expectedResponse:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:               "ERROR: Invalid cancellation reason",
			paymentID:          "payment-456",
			cancellationReason: "INVALID_REASON",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "INVALID_REASON"},
			expectedCode:       http.StatusBadRequest,
			expectedResponse:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'CancelUnifiedPaymentSessionRequest.CancellationReason' Error:Field validation for 'CancellationReason' failed on the 'oneof' tag"}],"traceId":""}}`,
		},
		{
			name:               "ERROR: Service returns error",
			paymentID:          "payment-456",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:       http.StatusBadRequest,
			expectedResponse:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"payment cannot be cancelled"}],"traceId":""}}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "payment-456" &&
						req.MerchantID == "merchant-123" &&
						req.CancellationReason == "REQUESTED_BY_CUSTOMER" &&
						req.Source == "MERCHANT"
				})).Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("payment cannot be cancelled"))).Once()
			},
		},
		{
			name:               "SUCCESS: Cancel payment",
			paymentID:          "payment-456",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:       http.StatusOK,
			expectedResponse:   `{"code":"00","data":{"amount":{"currency":"","value":0},"autoConfirm":false,"bypassStatusPage":false,"chargeDetails":null,"clientReferenceId":"","createdAt":"0001-01-01T00:00:00Z","expiryAt":null,"id":"","mode":"","paymentMethod":null,"paymentType":"","redirectUrl":{"expirationReturnUrl":"","failureReturnUrl":"","successReturnUrl":""},"statementDescriptor":"","status":"CANCELLED","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "payment-456" &&
						req.MerchantID == "merchant-123" &&
						req.CancellationReason == "REQUESTED_BY_CUSTOMER" &&
						req.Source == "MERCHANT"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					Status: constant.UnifiedPaymentSessionStatusCancelled,
				}, nil).Once()
			},
		},
		{
			name:               "SUCCESS: With sub-merchant ID",
			paymentID:          "payment-456",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			requestHeader:      map[string]string{constant.HeaderXSubMerchantID: "sub-merchant-789"},
			expectedCode:       http.StatusOK,
			expectedResponse:   `{"code":"00","data":{"amount":{"currency":"","value":0},"autoConfirm":false,"bypassStatusPage":false,"chargeDetails":null,"clientReferenceId":"","createdAt":"0001-01-01T00:00:00Z","expiryAt":null,"id":"","mode":"","paymentMethod":null,"paymentType":"","redirectUrl":{"expirationReturnUrl":"","failureReturnUrl":"","successReturnUrl":""},"statementDescriptor":"","status":"CANCELLED","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "payment-456" &&
						req.MerchantID == "sub-merchant-789" &&
						req.CancellationReason == "REQUESTED_BY_CUSTOMER" &&
						req.Source == "MERCHANT"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					Status: constant.UnifiedPaymentSessionStatusCancelled,
				}, nil).Once()
			},
		},
		{
			name:               "SUCCESS: With custom source",
			paymentID:          "payment-456",
			cancellationReason: "REQUESTED_BY_CUSTOMER",
			merchantClaim:      &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:        map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER", "source": "CUSTOMER"},
			expectedCode:       http.StatusOK,
			expectedResponse:   `{"code":"00","data":{"amount":{"currency":"","value":0},"autoConfirm":false,"bypassStatusPage":false,"chargeDetails":null,"clientReferenceId":"","createdAt":"0001-01-01T00:00:00Z","expiryAt":null,"id":"","mode":"","paymentMethod":null,"paymentType":"","redirectUrl":{"expirationReturnUrl":"","failureReturnUrl":"","successReturnUrl":""},"statementDescriptor":"","status":"CANCELLED","updatedAt":"0001-01-01T00:00:00Z"},"message":"Success"}`,
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "payment-456" &&
						req.MerchantID == "merchant-123" &&
						req.CancellationReason == "REQUESTED_BY_CUSTOMER" &&
						req.Source == "CUSTOMER"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					Status: constant.UnifiedPaymentSessionStatusCancelled,
				}, nil).Once()
			},
		},
		{
			name:          "ERROR: payment not found should return 422",
			paymentID:     "payment-456",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:   map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:  http.StatusUnprocessableEntity,
			wantContains:  "payment not found",
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.CancelUnifiedPaymentSessionRequest")).
					Return(nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)).Once()
			},
		},
		{
			name:          "ERROR: merchant mismatch should return 422",
			paymentID:     "payment-456",
			merchantClaim: &merchant.MerchantAuthTokenClaims{MerchantId: "merchant-123"},
			requestBody:   map[string]interface{}{"cancellationReason": "REQUESTED_BY_CUSTOMER"},
			expectedCode:  http.StatusUnprocessableEntity,
			wantContains:  "merchant id is not match",
			setup: func(mockService *serviceMocks.IUnifiedPaymentService) {
				mockService.On("CancelSession", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.CancelUnifiedPaymentSessionRequest")).
					Return(nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)).Once()
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
				reqBody, err = json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			} else if tc.name == "ERROR: Invalid JSON" {
				// Send invalid JSON for this specific test case
				reqBody = []byte("invalid json {{{")
			}

			req := httptest.NewRequest(http.MethodPost, "/payments/"+tc.paymentID+"/cancel", bytes.NewBuffer(reqBody))
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

			controller.Cancel(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Result().StatusCode)
			if tc.wantContains != "" {
				assert.Contains(t, rec.Body.String(), tc.wantContains)
			} else {
				assert.JSONEq(t, tc.expectedResponse, rec.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
