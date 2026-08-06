package v1CrmPaymentController_test

import (
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"
)

func TestGetDetailByID(t *testing.T) {
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
				svc.On("GetDetailByID", mock.Anything, validPaymentID).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:      "SUCCESS",
			paymentID: validPaymentID,
			mockService: func() {
				svc.On("GetDetailByID", mock.Anything, validPaymentID).
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/payments/%s", test.paymentID), nil)

			router := chi.NewRouter()
			router.Get("/payments/{id}", h.GetDetailByID)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestGetSplitRoutingByTransferID(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validPaymentID := uuid.NewString()
	validTransferID := uuid.NewString()

	tests := []struct {
		name           string
		paymentID      string
		transferID     string
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid paymentID format",
			paymentID:      "invalid",
			transferID:     validTransferID,
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:           "ERROR: Invalid transferID format",
			paymentID:      validPaymentID,
			transferID:     "invalid",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"split routing payment id is not valid"}`,
		},
		{
			name:       "ERROR: Service error",
			paymentID:  validPaymentID,
			transferID: validTransferID,
			mockService: func() {
				svc.On("GetSplitRoutingByTransferID", mock.Anything, validPaymentID, validTransferID).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "SUCCESS",
			paymentID:  validPaymentID,
			transferID: validTransferID,
			mockService: func() {
				svc.On("GetSplitRoutingByTransferID", mock.Anything, validPaymentID, validTransferID).
					Once().Return(&splitRoutingPaymentModel.SplitRoutingPaymentDetailResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"amount":0, "clientReferenceId":"", "createdAt":"0001-01-01T00:00:00Z", "currency":"", "destinationMerchantId":"", "paymentId":"", "remarks":"", "sourceMerchantId":"", "transferId":"", "updatedAt":"0001-01-01T00:00:00Z"}, "message":"OK"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/payments/%s/split-routing/%s", test.paymentID, test.transferID), nil)

			router := chi.NewRouter()
			router.Get("/payments/{paymentId}/split-routing/{transferId}", h.GetSplitRoutingByTransferID)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
