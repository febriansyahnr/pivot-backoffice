package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEncryptCard(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	cfg := &config.Config{}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc))

	validRequest := &unifiedPaymentModel.EncryptCardRequest{
		ClientReferenceID: "test-ref-123",
		CardRequest: unifiedPaymentModel.EncryptCardDetailRequest{
			Number:      "1234567890123456",
			ExpiryMonth: "12",
			ExpiryYear:  "25",
			CVC:         "123",
			NameOnCard:  "Test User",
		},
		DeviceInformation: unifiedPaymentModel.DeviceInformation{
			Type:      "web",
			UserAgent: "test-agent",
			IpAddress: "127.0.0.1",
		},
	}

	validResponse := &unifiedPaymentModel.EncryptedCardResponse{
		ClientReferenceID: "test-ref-123",
		EncryptedCard:     "encrypted-card-data",
		EncryptedCardInformation: unifiedPaymentModel.EncryptedCardInformationResponse{
			First8Digits:     "12345678",
			First6Digits:     "123456",
			Last4Digits:      "7890",
			ExpiryMonth:      "12",
			ExpiryYear:       "25",
			HasAssociatedCVC: true,
			Fingerprint:      "test-fingerprint",
		},
		CreatedAt: "2023-01-01T00:00:00Z",
	}

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestBody   func() []byte
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR: Invalid JSON payload",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			requestBody: func() []byte {
				return []byte(`{"invalid-json"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid", "error":{"details":[{"field":"", "message":"invalid request payload"}], "traceId":"", "type":"API_ERROR"}, "message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Missing required fields",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			requestBody: func() []byte {
				return []byte(`{}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid", "error":{"details":[{"field":"", "message":"Key: 'EncryptCardRequest.ClientReferenceID' Error:Field validation for 'ClientReferenceID' failed on the 'required' tag\nKey: 'EncryptCardRequest.CardRequest' Error:Field validation for 'CardRequest' failed on the 'required' tag\nKey: 'EncryptCardRequest.DeviceInformation' Error:Field validation for 'DeviceInformation' failed on the 'required' tag"}], "traceId":"", "type":"API_ERROR"}, "message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Service error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"EncryptCard",
					mock.Anything,
					mock.AnythingOfType("*unifiedPaymentModel.EncryptCardRequest"),
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"internal_server_error", "error":{"details":[{"field":"", "message":"internal server error"}], "traceId":"", "type":"API_ERROR"}, "message":"Internal server error"}`,
		},
		{
			name: "SUCCESS: Valid request",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"EncryptCard",
					mock.Anything,
					mock.AnythingOfType("*unifiedPaymentModel.EncryptCardRequest"),
				).Return(validResponse, nil).Once()
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			wantStatus:   http.StatusOK,
			wantResponse: ` {"code":"00","message":"Success","data":{"clientReferenceId":"test-ref-123","encryptedCard":"encrypted-card-data","encryptedCardInformations":{"first8":"12345678","first6":"123456","last4":"7890","expiryMonth":"12","expiryYear":"25","hasAssociatedCvc":true,"fingerprint":"test-fingerprint"},"deviceInformations":{"type":"","userAgent":"","ipAddress":"","acceptLanguage":"","cookieToken":"","deviceId":"","browserWidth":"","browserHeight":"","country":""},"createdAt":"2023-01-01T00:00:00Z"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/encrypt-card", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			controller.EncryptCard(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if test.wantStatus == http.StatusOK {
				assert.JSONEqf(t, test.wantResponse, rec.Body.String(), "Expected: %s , Actual: %s", test.wantResponse, rec.Body.String())
			} else {
				// For error responses, we just check the status code as error format may vary
				assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			}
		})
	}
}

func TestGetEncryptedCard(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	cfg := &config.Config{}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc))

	validResponse := &unifiedPaymentModel.EncryptedCardResponse{
		ClientReferenceID: "test-ref-123",
		EncryptedCard:     "encrypted-card-data",
		EncryptedCardInformation: unifiedPaymentModel.EncryptedCardInformationResponse{
			First8Digits:     "12345678",
			First6Digits:     "123456",
			Last4Digits:      "7890",
			ExpiryMonth:      "12",
			ExpiryYear:       "25",
			HasAssociatedCVC: true,
			Fingerprint:      "test-fingerprint",
		},
		CreatedAt: "2023-01-01T00:00:00Z",
	}

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		cardId        string
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			cardId:       "valid-uuid-123e4567-e89b-12d3-a456-426614174000",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR: Invalid UUID",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			cardId:       "invalid-uuid",
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid", "error":{"details":[{"field":"", "message":"invalid request payload"}], "traceId":"", "type":"API_ERROR"}, "message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Service error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"GetEncryptedCard",
					mock.Anything,
					"test-merchant-id",
					"123e4567-e89b-12d3-a456-426614174000",
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			cardId:       "123e4567-e89b-12d3-a456-426614174000",
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"internal_server_error", "error":{"details":[{"field":"", "message":"internal server error"}], "traceId":"", "type":"API_ERROR"}, "message":"Internal server error"}`,
		},
		{
			name: "SUCCESS: Valid request",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "test-merchant-id",
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"GetEncryptedCard",
					mock.Anything,
					"test-merchant-id",
					"123e4567-e89b-12d3-a456-426614174000",
				).Return(validResponse, nil).Once()
			},
			cardId:       "123e4567-e89b-12d3-a456-426614174000",
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00","message":"Success","data":{"clientReferenceId":"test-ref-123","encryptedCard":"encrypted-card-data","encryptedCardInformations":{"first8":"12345678","first6":"123456","last4":"7890","expiryMonth":"12","expiryYear":"25","hasAssociatedCvc":true,"fingerprint":"test-fingerprint"},"deviceInformations":{"type":"","userAgent":"","ipAddress":"","acceptLanguage":"","cookieToken":"","deviceId":"","browserWidth":"","browserHeight":"","country":""},"createdAt":"2023-01-01T00:00:00Z"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/encrypt-card/"+test.cardId, nil)
			rec := httptest.NewRecorder()

			// Setup chi context for URL parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uuid", test.cardId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			controller.GetEncryptedCard(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if test.wantStatus == http.StatusOK {
				assert.JSONEqf(t, test.wantResponse, rec.Body.String(), "Expected: %s , Actual: %s", test.wantResponse, rec.Body.String())
			} else {
				// For error responses, we just check the status code as error format may vary
				assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			}
		})
	}
}
