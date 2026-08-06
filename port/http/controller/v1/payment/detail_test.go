package payment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestGetPaymentDetailForPaymentUI(t *testing.T) {
	paymentService := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validator.New(), nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/payments/detail", handler.GetPaymentDetailForPaymentUI)

	tests := []struct {
		name           string
		paymentID      string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid token",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "invalid token"),
		},
		{
			name:      "ERROR:Some error",
			paymentID: "payment-123",
			setupMock: func() {
				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS",
			paymentID: "payment-123",
			setupMock: func() {
				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Return(&paymentModel.PaymentDetailForPaymentUIResponse{
					UUID: "payment-123",
					Amount: commonModel.Amount{
						Value:    "100000.00",
						Currency: "IDR",
					},
					Status: "pending",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"payment-123","merchantId":"","customerId":"","referenceId":"","merchant":{"name":"","logo":""},"paymentMethod":{"name":"","logo":null,"method":"","category":""},"paymentTypeDetail":{},"processorReferenceNumber":"","paymentChannel":"","bankReferenceId":"","amount":{"currency":"IDR","value":"100000.00"},"amountPaid":{"currency":"","value":""},"status":"pending","transactionId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","expiredAt":null,"paidAt":null,"redirectUrl":{"successUrl":"","failedUrl":""},"bypassStatusPage":false}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/detail", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.paymentID != "" {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxPaymentID, test.paymentID))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetPaymentImages(t *testing.T) {
	paymentService := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validator.New(), nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/payments/images", handler.GetPaymentImages)

	tests := []struct {
		name           string
		paymentID      string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid token",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "invalid token"),
		},
		{
			name:      "ERROR:Some error",
			paymentID: "payment-123",
			setupMock: func() {
				paymentService.On(
					"GetImages", c.ValueCtxMockType(),
				).Once().Return(paymentModel.ImageResponse{}, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS",
			paymentID: "payment-123",
			setupMock: func() {
				paymentService.On(
					"GetImages", c.ValueCtxMockType(),
				).Return(paymentModel.ImageResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"securedImages":null,"poweredImages":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/images", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.paymentID != "" {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxPaymentID, test.paymentID))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetPaymentInstructions(t *testing.T) {
	paymentService := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validator.New(), nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/payments/instructions", handler.GetPaymentInstructions)

	tests := []struct {
		name           string
		paymentID      string
		paymentMethod  string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid token",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "invalid token"),
		},
		{
			name:          "ERROR:Some error",
			paymentID:     "payment-123",
			paymentMethod: "credit_card",
			setupMock: func() {
				paymentService.On(
					"GetPaymentInstructions", c.ValueCtxMockType(), "credit_card",
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:          "SUCCESS",
			paymentID:     "payment-123",
			paymentMethod: "credit_card",
			setupMock: func() {
				paymentService.On(
					"GetPaymentInstructions", c.ValueCtxMockType(), "credit_card",
				).Return([]paymentModel.InstructionResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			url := "/payments/instructions"
			if test.paymentMethod != "" {
				url += "?paymentMethod=" + test.paymentMethod
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.paymentID != "" {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxPaymentID, test.paymentID))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
