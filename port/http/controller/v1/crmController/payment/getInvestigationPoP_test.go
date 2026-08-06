package v1CrmPaymentController_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigationProofOfPayment(t *testing.T) {
	service := serviceMocks.NewIPaymentService(t)

	handler := New(service)

	route := chi.NewRouter()
	route.Get(
		"/investigations/{paymentId}/proof-of-payment",
		handler.GetInvestigationProofOfPayment,
	)

	paymentID := "a616f5bd-aa97-4cbb-a4d7-a3e765c69fae"
	request := model.GetInvestigationProofOfPaymentRequest{
		PaymentID: paymentID,
	}
	response := &model.GetInvestigationProofOfPaymentResponse{
		SignedURL:     "https://",
		ExpiresAt:     time.Now().UTC(),
		MerchantNotes: "Test", // NOSONAR
	}
	responseJSON, _ := json.Marshal(response)

	tests := []struct {
		name             string
		paymentID        string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:Invalid payment ID format", // NOSONAR
			paymentID:        "ABC",                             // NOSONAR
			setupMock:        func() { /* Empty Function */ },
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","errors":"invalid payment ID format"}`, // NOSONAR
		},
		{
			name:      "ERROR:Some error", // NOSONAR
			paymentID: paymentID,
			setupMock: func() {
				service.On(
					"GetInvestigationProofOfPayment", mock.Anything, request,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","errors":"assert.AnError general error for testing"}`,
		},
		{
			name:      "SUCCESS", // NOSONAR
			paymentID: paymentID,
			setupMock: func() {
				service.On(
					"GetInvestigationProofOfPayment", mock.Anything, request,
				).Once().Return(response, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: fmt.Sprintf(`{"code":"00","message":"OK","data": %s}`, string(responseJSON)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/investigations/%s/proof-of-payment", tt.paymentID), nil)

			tt.setupMock()

			route.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Response Actual:", rec.Body.String())
			}
		})
	}
}
