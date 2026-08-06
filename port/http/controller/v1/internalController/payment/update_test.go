package internalPaymentController

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInternalPaymentController_Update(t *testing.T) {
	expired := time.Now().Add(1 * time.Hour)
	existedMerchant := &merchant.Merchant{
		UUID:          "merchant-merchant-id",
		Name:          "test",
		Description:   "test",
		Logo:          "test.png",
		MerchantEmail: "test@gmail.com",
		MerchantPhone: "081231231",
		PICEmail:      "testing@paper.id",
		PICPhone:      "08812321321",
		MID:           sql.NullString{String: "12341234", Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	testCases := []struct {
		name           string
		paymentId      string
		mockSetup      func(mockService *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMq *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		setupContext   func(ctx context.Context) context.Context
		setHeaders     func(req *http.Request)
		expectedStatus int
	}{
		{
			name:      "SUCCESS: Update payment",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(existedMerchant, nil)

				payment.On(
					"UpdatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentUpdateRequest"),
				).Return(&paymentModel.PaymentResponse{}, nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)

			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "SUCCESS: Update payment in behalf of submerchant",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(existedMerchant, nil)

				payment.On(
					"UpdatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentUpdateRequest"),
				).Return(&paymentModel.PaymentResponse{}, nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)

			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "FAILED: id param is empty",
			paymentId: "",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR: Invalid JSON",
			paymentId:      "uuid-uuid-uuid",
			expectedStatus: http.StatusBadRequest,
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
		},
		{
			name:      "ERROR: Missing required request",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
		},
		{
			name:      "ERROR: Service error",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(existedMerchant, nil)

				payment.On(
					"UpdatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentUpdateRequest"),
				).Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error")))
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "FAILED: error when find merchant by id",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, assert.AnError)
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "FAILED: merchant info is nil",
			paymentId: "uuid-uuid-uuid",
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "ERROR: Merchant not in Context",
			paymentId: "uuid-uuid-uuid",
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentUpdateRequest{
					PaymentId:  "uuid-uuid-uuid",
					MerchantId: "merchant-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup: func(payment *serviceMocks.IPaymentService, merchant *serviceMocks.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {

			},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIPaymentService(t)
			mockMerchant := serviceMocks.NewIMerchantService(t)
			mockValidator := validatorExt.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tc.mockSetup(mockService, mockMerchant, mockRmq)
			paymentController := New(mockValidator, mockService, mockMerchant, mockRmq)
			chiRouterCtx := chi.NewRouteContext()

			baseUrl := "/api/internal/v1/payments/create"

			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tc.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(tc.setupContext(req.Context()))

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			chiRouterCtx.URLParams.Add("id", tc.paymentId)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(paymentController.Update)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSNAPUpdateVA(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	mockIP := "1.2.3.4" // NOSONAR
	service := serviceMocks.NewIPaymentService(t)
	rabbitMqExt := mockRabbitMq.NewRabbitMQExt(t)
	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)

	monitor.SetGlobalMonitoring(mntr)

	controller := &InternalPaymentController{
		paymentSvc:  service,
		logger:      logger,
		rabbitMqExt: rabbitMqExt,
	}

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/open-api/snap/v1/payments/transfer-va/update-va", controller.SNAPUpdateVA)

	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
		ClientId:   uuid.NewString(),
	}
	validBody := `{"trxId": "2b4510f3-6e27-4d64-9853-7e350a06479d", "virtualAccountName": "John Doe", "virtualAccountTrxType": "CLOSED_DYNAMIC", "totalAmount": {"value": "10000.00", "currency": "IDR"}, "expiredDate": "2024-08-28T13:59:59+07:00"}`
	invalidBody := `{"trxId": "2b4510f3-6e27-4d64-9853-7e350a06479d", "virtualAccountName": 12345}` // Invalid format for `virtualAccountName`

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
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4012801", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR: Invalid field format in request body",
			merchantAuth:   merchantAuth,
			bodyRequest:    invalidBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4002801", "Invalid Field Format virtualAccountName"),
		},
		{
			name:         "ERROR: Internal Server Error",
			merchantAuth: merchantAuth,
			bodyRequest:  validBody,
			setupMock: func() {
				service.On(
					"FindPaymentById", constant.ValueCtxMockType(), constant.StringMockType(), merchantAuth.MerchantId,
				).Once().Return(nil, pkgErrors.New(response.HttpErrNotFound, errors.New("Internal Server Error")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5002801", "Internal Server Error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/open-api/snap/v1/payments/transfer-va/update-va", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/transfer-va/update-va")
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
