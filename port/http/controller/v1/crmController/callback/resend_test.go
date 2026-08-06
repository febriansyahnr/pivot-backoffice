package crmCallbackController

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callback_model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResendCallback(t *testing.T) {
	logger, _ := logger.NewZapLogger(logger.Config{})
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	validMerchantID := uuid.NewString()
	validReferenceID := uuid.NewString()

	tests := []struct {
		name           string
		setupBody      func(*testing.T) []byte
		mockServices   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid JSON body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid request payload"}`,
		},
		{
			name: "ERROR: Missing merchantId",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					Type:              constant.TypePayment,
					ClientReferenceID: "test-ref-123",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantID":"Key: 'ResendCallbackRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Missing type",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					ClientReferenceID: "test-ref-123",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Type":"Key: 'ResendCallbackRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Invalid type value",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              "INVALID_TYPE",
					ClientReferenceID: "test-ref-123",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Type":"Key: 'ResendCallbackRequest.Type' Error:Field validation for 'Type' failed on the 'oneof' tag"}}`,
		},
		{
			name: "ERROR: Missing both clientReferenceId and referenceId",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID: validMerchantID,
					Type:       constant.TypePayment,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"ClientReferenceID":"Key: 'ResendCallbackRequest.ClientReferenceID' Error:Field validation for 'ClientReferenceID' failed on the 'required_without' tag"}}`,
		},
		{
			name: "ERROR: Invalid UUID format for referenceId",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              constant.TypePayment,
					ClientReferenceID: "test-ref-123",
					ReferenceID:       "invalid-uuid",
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				// no mocks needed
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"ReferenceID":"Key: 'ResendCallbackRequest.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'uuid' tag"}}`,
		},
		{
			name: "ERROR: Payment callback service error",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              constant.TypePayment,
					ClientReferenceID: "test-ref-123",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				unifiedPaymentSvc.On(
					"ResendPaymentCallback",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*callback_model.ResendCallbackRequest"),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "ERROR: Disbursement callback service error",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              constant.TypeDisbursement,
					ClientReferenceID: "test-ref-456",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				disbursementSvc.On(
					"ResendDisbursementCallback",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*callback_model.ResendCallbackRequest"),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS: Resend payment callback",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              constant.TypePayment,
					ClientReferenceID: "test-ref-123",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				unifiedPaymentSvc.On(
					"ResendPaymentCallback",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*callback_model.ResendCallbackRequest"),
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"clientReferenceId":"test-ref-123", "message":"Callback resent successfully","type":"PAYMENT","referenceId":"` + validReferenceID + `"}, "message": "OK"}`,
		},
		{
			name: "SUCCESS: Resend disbursement callback",
			setupBody: func(t *testing.T) []byte {
				payload := callback_model.ResendCallbackRequest{
					MerchantID:        validMerchantID,
					Type:              constant.TypeDisbursement,
					ClientReferenceID: "test-ref-456",
					ReferenceID:       validReferenceID,
				}
				payloadBytes, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadBytes
			},
			mockServices: func() {
				disbursementSvc.On(
					"ResendDisbursementCallback",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*callback_model.ResendCallbackRequest"),
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"clientReferenceId":"test-ref-456", "message":"Callback resent successfully","type":"DISBURSEMENT","referenceId":"` + validReferenceID + `"}, "message": "OK"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockServices()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/callback/resend", bytes.NewBuffer(test.setupBody(t)))

			handler := New(logger, unifiedPaymentSvc, disbursementSvc)
			handler.ResendCallback(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
