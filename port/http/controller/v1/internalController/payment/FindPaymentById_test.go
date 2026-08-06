package internalPaymentController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentController_FindPaymentById(t *testing.T) {
	paymentId, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err)

	paymentResponse := &paymentModel.PaymentResponse{
		UUID: uuid.NewString(),
	}

	testCase := []struct {
		name           string
		paymentId      string
		mockSetup      func(paymentSvc *mocks.IPaymentService)
		setupContext   func(ctx context.Context) context.Context
		setHeaders     func(req *http.Request)
		expectedStatus int
	}{
		{
			name:      "SUCCESS",
			paymentId: paymentId.String(),
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.On("FindPaymentById",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(paymentResponse, nil)
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: 200,
		},
		{
			name:      "SUCCESS in behalf of submerchant",
			paymentId: paymentId.String(),
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.On("FindPaymentById",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(paymentResponse, nil)
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			expectedStatus: 200,
		},
		{
			name:      "ERROR: Service Error",
			paymentId: paymentId.String(),
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.On("FindPaymentById",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("service error"))
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "ERROR: Payment id is required",
			paymentId: "",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "ERROR: Merchant Auth not in Context",
			paymentId: paymentId.String(),
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			paymentSvc := mocks.NewIPaymentService(t)
			merchantSvc := mocks.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(paymentSvc)

			svc := New(mockValidator, paymentSvc, merchantSvc, mockRmq)
			chiRouterCtx := chi.NewRouteContext()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/internal/v1/payments", nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(tt.setupContext(req.Context()))

			if tt.setHeaders != nil {
				tt.setHeaders(req)
			}

			chiRouterCtx.URLParams.Add("id", tt.paymentId)

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(svc.FindPaymentById)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			paymentSvc.AssertExpectations(t)
		})
	}
}

func TestSNAPGetVA(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	mockIP := "1.2.3.4" // NOSONAR
	service := mocks.NewIPaymentService(t)

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)

	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/payments/transfer-va/get-va", New(validator.New(), service, nil, nil, WithLogger(logger)).SNAPGetVA)
	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
		ClientId:   uuid.NewString(),
	}
	bodyRequest := `{"trxId": "2b4510f3-6e27-4d64-9853-7e350a06479d",  "virtualAccountNo" : "76630118700"}`
	trxIdMockType := constant.StringMockType()

	tests := []struct {
		name           string
		bodyRequest    string
		merchantAuth   *merchant.MerchantAuthTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Merchant auth not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4013001", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR: Invalid field format trxId",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"trxId": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4003001", "Invalid Field Format trxId"),
		},
		{
			name:         "ERROR: Transaction not found",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"FindPaymentById", constant.ValueCtxMockType(), trxIdMockType, constant.StringMockType(),
				).Once().Return(nil, pkgErrs.New(response.HttpErrNotFound, pkgErrs.New("data not found", nil)))
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4043012", "Invalid Bill/Virtual Account"),
		},
		{
			name:         "ERROR: Internal error",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"FindPaymentById", constant.ValueCtxMockType(), trxIdMockType, constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5003001", "Internal Server Error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log("test.bodyRequest,", test.bodyRequest)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/payments/transfer-va/get-va", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/transfer-va/get-va")
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, test.merchantAuth))
			}

			// Check if the router and dependencies are correctly initialized
			require.NotNil(t, router)
			require.NotNil(t, service)
			require.NotNil(t, logger)

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
