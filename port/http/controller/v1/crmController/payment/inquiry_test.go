package v1CrmPaymentController

import (
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

func TestInquiryByID(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validPaymentID := uuid.NewString()
	tests := []struct {
		name           string
		paymentID      string
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid paymentID format",
			paymentID:      "invalid",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:      "ERROR: Service error",
			paymentID: validPaymentID,
			mockService: func() {
				svc.On("InquiryPayment", mock.Anything, mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:      "SUCCESS",
			paymentID: validPaymentID,
			mockService: func() {
				svc.On("InquiryPayment", mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"amount":"0", "createdAt":"0001-01-01T00:00:00Z", "currency":"", "customerId":"", "deletedAt":null, "discount":null, "expiredAt":null, "fee":null, "investigationStartedAt":null, "merchantId":"", "metadata":null, "paymentMethodId":"", "paymentUrl":"", "processorReferenceNumber":null, "referenceId":null, "status":"", "totalAmount":"0", "type": "", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}, "message":"OK"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments/%s/inquiry", test.paymentID), nil)

			router := chi.NewRouter()
			router.Post("/payments/{id}/inquiry", h.InquiryByID)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
