package v1CrmPaymentController

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetryStaticVANotification(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validPayload := paymentModel.CRMStaticVARetryNotificationRequest{
		VANumber: "88012345678",
		Amount: commonModel.Amount{
			Value:    "100000",
			Currency: "IDR",
		},
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid JSON payload",
			requestBody:    "invalid-json",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name: "ERROR: Missing vaNumber",
			requestBody: paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: "",
				Amount: commonModel.Amount{
					Value:    "100000",
					Currency: "IDR",
				},
			},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"VANumber":"Key: 'CRMStaticVARetryNotificationRequest.VANumber' Error:Field validation for 'VANumber' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Missing amount",
			requestBody: paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: "88012345678",
			},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Amount":"Key: 'CRMStaticVARetryNotificationRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: validPayload,
			mockService: func() {
				svc.On("CRMStaticVARetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMStaticVARetryNotificationRequest) bool {
					return req.VANumber == "88012345678" && req.Amount.Value == "100000"
				})).Once().Return(pkgErrors.New(response.HttpErrInternal, errors.New("snap core error")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"snap core error"}`,
		},
		{
			name:        "SUCCESS: Valid request",
			requestBody: validPayload,
			mockService: func() {
				svc.On("CRMStaticVARetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMStaticVARetryNotificationRequest) bool {
					return req.VANumber == "88012345678" && req.Amount.Value == "100000" && req.Amount.Currency == "IDR"
				})).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"message":"Static VA payment notification published successfully","vaNumber":"88012345678"},"message":"OK"}`,
		},
		{
			name: "SUCCESS: Different VA number and amount",
			requestBody: paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: "88098765432",
				Amount: commonModel.Amount{
					Value:    "250000",
					Currency: "IDR",
				},
			},
			mockService: func() {
				svc.On("CRMStaticVARetryNotification", mock.Anything, mock.MatchedBy(func(req *paymentModel.CRMStaticVARetryNotificationRequest) bool {
					return req.VANumber == "88098765432" && req.Amount.Value == "250000"
				})).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"message":"Static VA payment notification published successfully","vaNumber":"88098765432"},"message":"OK"}`,
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
			req := httptest.NewRequest(http.MethodPost, "/payments/static-va/retry-payment-notif", bytes.NewReader(body))

			router := chi.NewRouter()
			router.Post("/payments/static-va/retry-payment-notif", h.RetryStaticVANotification)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
