package simulationController

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestGetPaymentByID(t *testing.T) {
	paymentSvc := serviceMocks.NewIPaymentService(t)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
	handler := New(validatorExt.New(), WithPaymentService(paymentSvc), WithPaymentMethodService(paymentMethodSvc))

	validPaymentID := uuid.NewString()

	testCases := []struct {
		name           string
		paymentID      string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:      "ERROR: Invalid id",
			paymentID: "invalid",
			setupMock: func() {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"id is required"}`,
		},
		{
			name:      "ERROR: FindPaymentForSimulationByID service error",
			paymentID: validPaymentID,
			setupMock: func() {
				paymentSvc.On("FindPaymentForSimulationByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:      "SUCCESS",
			paymentID: validPaymentID,
			setupMock: func() {
				paymentSvc.On("FindPaymentForSimulationByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.PaymentResponse{UUID: "mock-uuid"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"merchantId":"", "referenceId":"", "status":"", "uuid":"mock-uuid"},"message":"OK"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/simulations/payment/"+tc.paymentID, nil)

			router := chi.NewRouter()
			router.Get("/simulations/payment/{id}", handler.GetPaymentByID)
			tc.setupMock()

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})

	}
}

func TestProcessPayment(t *testing.T) {
	paymentSvc := serviceMocks.NewIPaymentService(t)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
	handler := New(validatorExt.New(), WithPaymentService(paymentSvc), WithPaymentMethodService(paymentMethodSvc))

	validPaymentID := uuid.NewString()

	payload := &paymentModel.ProcessPaymentSimulation{
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "100000.00",
		},
	}

	testCases := []struct {
		name           string
		paymentID      string
		setupMock      func()
		request        string
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:      "ERROR: Invalid id",
			paymentID: "invalid",
			setupMock: func() {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"id is required"}`,
		},
		{
			name:      "ERROR: Invalid request",
			paymentID: validPaymentID,
			request:   `{/}`,
			setupMock: func() {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid character '/' looking for beginning of object key string"}`,
		},
		{
			name:      "ERROR: Missing request",
			paymentID: validPaymentID,
			request:   `{"test":"test"}`,
			setupMock: func() {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","data":null,"error":{"details":[{"field":"PaidAmount","message":"Key: 'ProcessPaymentSimulation.PaidAmount' Error:Field validation for 'PaidAmount' failed on the 'required' tag"}],"traceId":"","type":"API_VALIDATION_ERROR"},"message":"invalid validation"}`,
		},
		{
			name:      "ERROR: ProcessPaymentForSimulationByID service error",
			paymentID: validPaymentID,
			setupMock: func() {
				paymentSvc.On("ProcessPaymentForSimulationByID", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("commonModel.Amount"), constant.StringMockType()).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:      "SUCCESS",
			paymentID: validPaymentID,
			setupMock: func() {
				paymentSvc.On("ProcessPaymentForSimulationByID", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("commonModel.Amount"), constant.StringMockType()).
					Once().Return(nil)
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.Payment{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"` + validPaymentID + `", "redirectUrl":""},"message":"OK"}`,
		},
		{
			name:      "Failed to get redirection url",
			paymentID: validPaymentID,
			setupMock: func() {
				paymentSvc.On("ProcessPaymentForSimulationByID", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("commonModel.Amount"), constant.StringMockType()).
					Once().Return(nil)
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrPaymentNotFound)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"payment not found"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if tc.request == "" {
				buf, _ := json.Marshal(payload)
				tc.request = string(buf)
			}

			req := httptest.NewRequest(http.MethodPost, "/simulations/payment/"+tc.paymentID, strings.NewReader(tc.request))

			router := chi.NewRouter()
			router.Post("/simulations/payment/{id}", handler.ProcessPayment)
			tc.setupMock()

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})

	}
}
