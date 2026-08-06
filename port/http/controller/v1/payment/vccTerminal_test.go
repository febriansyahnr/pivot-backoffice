package payment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVCCTerminalBatchCharge(t *testing.T) {
	const (
		merchantID = "550e8400-e29b-41d4-a716-446655440000"
		userID     = "550e8400-e29b-41d4-a716-446655440001"
		// Valid base64-encoded values required by DataEncryption validation
		requestBody = `{"encryptedKey":"dGVzdEtleQ==","nonce":"dGVzdE5vbmNl","ciphertext":"dGVzdENpcGhlcnRleHQ="}`
	)

	userInfo := &userModel.UserTokenClaims{
		UUID:       userID,
		MerchantId: merchantID,
	}
	paymentService := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validatorExt.New(), nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Post("/vcc-terminal/charges/batch", handler.VCCTerminalBatchCharge)

	tests := []struct {
		name             string
		userInfo         *userModel.UserTokenClaims
		requestBody      string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR: User not found in context", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid JSON request body", // NOSONAR
			userInfo:         userInfo,
			requestBody:      `invalid`,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid character 'i' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Mandatory fields", // NOSONAR
			userInfo:         userInfo,
			requestBody:      `{"encryptedKey":"MDAwMDAwMDAw", "nonce":"MDAwMDAwMDAw", "ciphertext":""}`,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Ciphertext","message":"Key: 'VCCTerminalChargeRequest.EncryptedRequest.Ciphertext' Error:Field validation for 'Ciphertext' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Some error", // NOSONAR
			userInfo:    userInfo,
			requestBody: requestBody,
			setupMock: func() {
				paymentService.On("VCCTerminalBatchCharge", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS", // NOSONAR
			userInfo:    userInfo,
			requestBody: requestBody,
			setupMock: func() {
				paymentService.On("VCCTerminalBatchCharge", mock.Anything, mock.Anything).Once().Return(&paymentModel.VCCTerminalBatchChargeResponse{
					SuccessCount: 1,
					SuccessTotal: 150_000,
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"batchId":"","successCount":1,"successTotal":150000,"failedCount":0,"failedTotal":0,"failedCharges":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/vcc-terminal/charges/batch", strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userInfo != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userInfo))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}
