package v1CrmPaymentController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validPaymentID := uuid.NewString()
	validPayload := paymentModel.CRMRetryNotificationRequest{
		BankReference: "BANK12345",
	}

	tests := []struct {
		name           string
		paymentID      string
		requestBody    interface{}
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid paymentID format",
			paymentID:      "invalid-uuid",
			requestBody:    validPayload,
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:           "ERROR: Empty paymentID",
			paymentID:      "",
			requestBody:    validPayload,
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:           "ERROR: Invalid JSON payload",
			paymentID:      validPaymentID,
			requestBody:    "invalid-json",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name:        "ERROR: Service error",
			paymentID:   validPaymentID,
			requestBody: validPayload,
			mockService: func() {
				svc.On("CRMRetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMRetryNotificationRequest) bool {
					return req.ID == validPaymentID && req.BankReference == "BANK12345"
				})).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:        "SUCCESS: Valid request",
			paymentID:   validPaymentID,
			requestBody: validPayload,
			mockService: func() {
				svc.On("CRMRetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMRetryNotificationRequest) bool {
					return req.ID == validPaymentID && req.BankReference == "BANK12345"
				})).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s","message":"Payment published to CRM successfully"},"message":"OK"}`, validPaymentID),
		},
		{
			name:      "SUCCESS: Empty bank reference",
			paymentID: validPaymentID,
			requestBody: paymentModel.CRMRetryNotificationRequest{
				BankReference: "",
			},
			mockService: func() {
				svc.On("CRMRetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMRetryNotificationRequest) bool {
					return req.ID == validPaymentID && req.BankReference == ""
				})).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s","message":"Payment published to CRM successfully"},"message":"OK"}`, validPaymentID),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			var body []byte
			if str, ok := test.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments/%s/retry-notification", test.paymentID), bytes.NewReader(body))

			router := chi.NewRouter()
			router.Post("/payments/{id}/retry-notification", h.RetryNotification)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
