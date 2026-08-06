package internalPaymentController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const VaNumber = "1234123412341234"

var customerID = uuid.NewString()
var validPayload = paymentModel.PaymentRequest{
	ReferenceID: uuid.NewString(),
	Customer: paymentModel.PaymentRequestCustomer{
		Name: "John Doe", CustomerID: customerID, Email: "a@a.com",
	},
	PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
	TotalAmount: paymentModel.Amount{
		Currency: "IDR", Value: decimal.NewFromInt(1000000),
	},
	VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
		Issuer:                constant.BANK_ACQUIRER_PERMATA,
		VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
		VirtualAccountName:    "John Doe",
		MinAmount: &paymentModel.Amount{
			Currency: "IDR", Value: decimal.NewFromInt(1000000),
		},
		MaxAmount: &paymentModel.Amount{
			Currency: "IDR", Value: decimal.NewFromInt(1000000),
		},
	},
	PaymentItems: &[]paymentModel.PaymentItemRequest{
		{
			Name: "nasi goreng",
			Qty:  1,
			Amount: paymentModel.Amount{
				Currency: "IDR", Value: decimal.NewFromInt(1000000),
			},
		},
	},
}

func TestCreate(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		merchantClaim  *merchant.MerchantAuthTokenClaims
		setHeaders     func(req *http.Request)
	}{
		{
			name: "SUCCESS: Create payments using VA",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"CreatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("paymentModel.PaymentRequest"),
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
				var paymentItemsRequest []paymentModel.PaymentItemRequest
				paymentItemsRequest = append(paymentItemsRequest, paymentModel.PaymentItemRequest{
					Name: "Bill for A",
					Qty:  1,
					Amount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					Metadata: &map[string]any{
						"BillCode": "INV-JAN-01",
						"BillName": "SPP JAN",
					},
				})

				payload := paymentModel.PaymentRequest{
					ReferenceID: uuid.NewString(),
					Customer: paymentModel.PaymentRequestCustomer{
						Name:       "John Doe",
						CustomerID: customerID,
						Email:      "a@a.com",
					},
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					PaymentItems: &paymentItemsRequest,
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						Issuer:                constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
						VirtualAccountName:    "John Doe",
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
					},
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchantClaim,
		},
		{
			name: "SUCCESS: Create payments using VA in behalf of submerchant",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"CreatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("paymentModel.PaymentRequest"),
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
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			setupBody: func(t *testing.T) []byte {
				var paymentItemsRequest []paymentModel.PaymentItemRequest
				paymentItemsRequest = append(paymentItemsRequest, paymentModel.PaymentItemRequest{
					Name: "Bill for A",
					Qty:  1,
					Amount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					Metadata: &map[string]any{
						"BillCode": "INV-JA	N-01",
						"BillName": "SPP JAN",
					},
				})

				payload := paymentModel.PaymentRequest{
					ReferenceID: uuid.NewString(),
					Customer: paymentModel.PaymentRequestCustomer{
						Name:       "John Doe",
						CustomerID: customerID,
						Email:      "a@a.com",
					},
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					PaymentItems: &paymentItemsRequest,
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						Issuer:                constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
						VirtualAccountName:    "John Doe",
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
					},
				}

				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchantClaim,
		},
		{
			name:           "ERROR: Invalid JSON",
			expectedStatus: http.StatusBadRequest,
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup:     func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			merchantClaim: validMerchantClaim,
		},
		{
			name:      "ERROR: Missing required request",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchantClaim,
		},
		{
			name:      "ERROR: Missing required paymentItems request",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				var paymentItemsRequest []paymentModel.PaymentItemRequest
				paymentItemsRequest = append(paymentItemsRequest, paymentModel.PaymentItemRequest{})

				payload := paymentModel.PaymentRequest{
					ReferenceID: uuid.NewString(),
					Customer: paymentModel.PaymentRequestCustomer{
						Name:       "John Doe",
						CustomerID: customerID,
						Email:      "a@a.com",
					},
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					PaymentItems: &paymentItemsRequest,
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						Issuer:                constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
						VirtualAccountNumber:  VaNumber,
						VirtualAccountName:    "John Doe",
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
					},
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchantClaim,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"CreatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("paymentModel.PaymentRequest"),
				).Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error")))
			},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{
					ReferenceID: uuid.NewString(),
					Customer: paymentModel.PaymentRequestCustomer{
						Name:       "John Doe",
						CustomerID: customerID,
						Email:      "a@a.com",
					},
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount: paymentModel.Amount{
						Value:    decimal.NewFromInt(1000000),
						Currency: "IDR",
					},
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						Issuer:                constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
						VirtualAccountName:    "John Doe",
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(1000000),
							Currency: "IDR",
						},
					},
					PaymentItems: &[]paymentModel.PaymentItemRequest{
						{
							Name: "nasi goreng",
							Qty:  1,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(1000000),
								Currency: "IDR",
							},
						},
					},
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusInternalServerError,
			merchantClaim:  validMerchantClaim,
		},
		{
			name:      "ERROR:Payment method is not allowed",
			mockSetup: func(*serviceMocks.IPaymentService, *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {

				payload := validPayload
				payload.PaymentMethod = "XXXX"

				buf, err := json.Marshal(payload)
				assert.NoError(t, err)
				return buf
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchantClaim,
		},
		{
			name:      "ERROR:Virtual account has expired",
			mockSetup: func(*serviceMocks.IPaymentService, *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {

				payload := validPayload
				payload.VirtualAccount.ExpiredDate = &time.Time{}

				buf, err := json.Marshal(payload)
				assert.NoError(t, err)
				return buf
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchantClaim,
		},
		{
			name: "ERROR: Merchant not in Context",
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup:      func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			merchantClaim:  nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIPaymentService(t)
			mockMerchant := serviceMocks.NewIMerchantService(t)
			mockValidator := validatorExt.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tc.mockSetup(mockService, mockRmq)
			paymentController := New(mockValidator, mockService, mockMerchant, mockRmq)

			baseUrl := "/api/internal/v1/payments/create"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tc.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(paymentController.Create)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestValidateVirtualAccountPayload(t *testing.T) {
	validPayload := paymentModel.PaymentRequest{
		ReferenceID:   uuid.NewString(),
		PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		TotalAmount: paymentModel.Amount{
			Value:    decimal.NewFromInt(1000000),
			Currency: "IDR",
		},
		Customer: paymentModel.PaymentRequestCustomer{
			Name:  "name",
			Email: "name@email.co",
			Phone: "081234567890",
		},
		PaymentItems: &[]paymentModel.PaymentItemRequest{
			{
				Name: "nasi goreng",
				Qty:  1,
				Amount: paymentModel.Amount{
					Value:    decimal.NewFromInt(1000000),
					Currency: "IDR",
				},
			},
		},
		VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
			VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
			Issuer:                constant.BANK_ACQUIRER_PERMATA,
			VirtualAccountName:    "validName",
		},
	}

	testCases := []struct {
		name     string
		payload  func() paymentModel.PaymentRequest
		expected error
	}{
		{
			name: "Nil virtual account",
			payload: func() paymentModel.PaymentRequest {
				nilVirtualAccount := validPayload
				nilVirtualAccount.VirtualAccount = nil
				return nilVirtualAccount
			},
			expected: pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccount object is required")),
		},
		{
			name: "Valid CLOSED_DYNAMIC payment",
			payload: func() paymentModel.PaymentRequest {
				return validPayload
			},
			expected: nil,
		},
		{
			name: "Valid OPEN_STATIC payment",
			payload: func() paymentModel.PaymentRequest {
				openStaticPayload := validPayload
				openStaticPayload.VirtualAccount.VirtualAccountTrxType = paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC
				openStaticPayload.VirtualAccount.VirtualAccountNumber = "76786789"

				return openStaticPayload
			},
			expected: nil,
		},
		{
			name: "Valid CLOSED_STATIC payment",
			payload: func() paymentModel.PaymentRequest {
				openStaticPayload := validPayload
				openStaticPayload.VirtualAccount.VirtualAccountTrxType = paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC

				return openStaticPayload
			},
			expected: nil,
		},
		{
			name: "Invalid virtual account type",
			payload: func() paymentModel.PaymentRequest {
				openStaticPayload := validPayload
				openStaticPayload.VirtualAccount.VirtualAccountTrxType = "INVALID_TYPE"

				return openStaticPayload
			},
			expected: pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccountTrxType is not allowed")),
		},
	}

	controller := InternalPaymentController{
		validate: validatorExt.New(),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := controller.validateVirtualAccountPayload(tc.payload())
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSNAPGenerateQRMpm(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	service := serviceMocks.NewIPaymentService(t)
	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)
	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/open-api/snap/v1/payments/qr/qr-mpm-generate", New(validator.New(), service, nil, nil, WithLogger(logger)).SNAPGenerateQRMpm)

	consistentMerchantId := uuid.NewString()
	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: consistentMerchantId,
	}

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
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4014701", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR: Invalid field format partnerReferenceNo",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": false,"subMerchantId": "merchant123", "validityPeriod": "3600", "amount": {"value": "1000.00", "currency": "IDR"}}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4004701", "Invalid Field Format partnerReferenceNo"),
		},
		{
			name:           "ERROR: Invalid field format subMerchantId",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "ref123","subMerchantId": false, "validityPeriod": "3600", "amount": {"value": "1000.00", "currency": "IDR"}}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4004701", "Invalid Field Format subMerchantId"),
		},
		{
			name:         "ERROR: Internal error",
			merchantAuth: merchantAuth,
			bodyRequest:  `{"partnerReferenceNo": "ref123", "additionalInfo": {"qrType": "STATIC"}}`,
			setupMock: func() {
				service.On(
					"CreatePayment", mock.Anything, consistentMerchantId, mock.AnythingOfType("paymentModel.PaymentRequest"),
				).Once().Return(nil, errors.New("some error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5004701", "Internal Server Error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/open-api/snap/v1/payments/qr/qr-mpm-generate", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXExternalId, "17185234321")
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/qr/qr-mpm-generate")
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, test.merchantAuth))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}

func TestSNAPCreateVA(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	service := serviceMocks.NewIPaymentService(t)
	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)
	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/open-api/snap/v1/payments/transfer-va/create-va", New(validator.New(), service, nil, nil, WithLogger(logger)).SNAPCreateVA)

	consistentMerchantId := uuid.NewString()
	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: consistentMerchantId,
	}

	tests := []struct {
		name           string
		bodyRequest    string
		merchantAuth   *merchant.MerchantAuthTokenClaims
		setupMock      func(ctx context.Context, req paymentModel.PaymentRequest)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Merchant auth not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4012701", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR: Invalid field format virtualAccountEmail",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"virtualAccountEmail": false, "virtualAccountPhone": "08123456789", "virtualAccountTrxType": "PAYMENT", "virtualAccountName": "John Doe"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4002701", "Invalid Field Format virtualAccountEmail"),
		},
		{
			name:           "ERROR: Invalid field format virtualAccountTrxType",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"virtualAccountEmail": "john.doe@example.com", "virtualAccountPhone": "08123456789", "virtualAccountTrxType": false, "virtualAccountName": "John Doe"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4002701", "Invalid Field Format virtualAccountTrxType"),
		},
		{
			name:           "ERROR: Internal Server",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"expiredDate":"2500-08-27T23:59:59-07:00", "virtualAccountEmail": "email@orang.com", "virtualAccountPhone": "08990092019", "virtualAccountTrxType": "CLOSED_DYNAMIC", "virtualAccountName": "AHMAD YUSUF ARDABILLI", "totalAmount": {"value": "10000.00", "currency": "IDR"}, "billDetails": [{"billName": "BAYAR AIR", "billDescription": {"english": "pay on water PAM", "indonesia": "bayar air pam"}, "billAmount": {"value": "10000.00", "currency": "IDR"}}], "additionalInfo": {"referenceId": "BAYAR-AIR-1", "customer": {"name": "customerName", "phone": "08990090909", "email": "customer@email.com"}, "issuer": "permata"}}`,
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5002701", "Internal Server Error"),
			setupMock: func(ctx context.Context, req paymentModel.PaymentRequest) {
				service.On(
					"CreatePayment", mock.Anything, consistentMerchantId, mock.AnythingOfType("paymentModel.PaymentRequest"),
				).Once().Return(nil, errors.New("some error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/open-api/snap/v1/payments/transfer-va/create-va", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/transfer-va/create-va")
			ctx := context.Background()
			if test.merchantAuth != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantAuth)
			}
			ctx = context.WithValue(ctx, constant.CtxSnapApiName, "27")
			req = req.WithContext(ctx)
			totalAmount, err := decimal.NewFromString("10000")
			if err != nil {
				t.Fatalf("Failed to parse total amount: %v", err)
			}
			paymentItemsRequest := []paymentModel.PaymentItemRequest{
				{
					Amount: paymentModel.Amount{
						Value:    totalAmount,
						Currency: "IDR",
					},
				},
			}

			parsedTime, err := time.Parse(time.RFC3339, "2024-08-27T23:59:59-07:00")
			if err != nil {
				t.Fatalf("Failed to parse time: %v", err)
			}

			expectedRequest := paymentModel.PaymentRequest{
				ReferenceID:   "BAYAR-AIR-1",
				PaymentMethod: "VIRTUAL_ACCOUNT",
				TotalAmount: paymentModel.Amount{
					Value:    totalAmount,
					Currency: "IDR",
				},
				VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
					Issuer:                "permata",
					VirtualAccountTrxType: "CLOSED_DYNAMIC",
					VirtualAccountName:    "AHMAD YUSUF ARDABILLI",
					ExpiredDate:           &parsedTime,
				},
				PaymentItems: &paymentItemsRequest,
				Customer: paymentModel.PaymentRequestCustomer{
					Name:  "customerName",
					Email: "customer@email.com",
					Phone: "08990090909",
				},
				IsSnap: true,
			}
			t.Log("req.Context():", test.bodyRequest)
			if test.setupMock != nil {
				test.setupMock(ctx, expectedRequest)
			}

			router.ServeHTTP(rec, req)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
