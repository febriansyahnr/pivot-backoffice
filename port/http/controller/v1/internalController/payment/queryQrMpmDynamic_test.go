package internalPaymentController

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestQueryQrMpmDynamic(t *testing.T) {
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	testCase := []struct {
		name           string
		uuid           string
		referenceId    string
		query          string
		reqSetting     func(r *http.Request)
		mockSetup      func(paymentSvc *mocks.IPaymentService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
	}{
		{
			name:  "success query qr mpm dynamic",
			query: "uuid=12345",
			mockSetup: func(mockPayment *mocks.IPaymentService) {
				mockPayment.On("GetQrMpmDynamic", mock.Anything, "12345", "", "12345").Return(&paymentModel.PaymentResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			reqSetting:     validRequestID,
			setHeaders: func(req *http.Request) {
				req.Header.Set("X-SubMerchantID", "12345")
			},
		},
		{
			name:           "error invalid merchant",
			query:          "uuid=12345",
			wantStatusCode: http.StatusUnauthorized,
			setHeaders: func(req *http.Request) {
				req.Header.Set("X-SubMerchantID", "12345")
			},
		},
		{
			name:           "error uuid and referenceId is required",
			query:          "",
			reqSetting:     validRequestID,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:  "error get qr mpm dynamic",
			query: "uuid=12345",
			mockSetup: func(mockPayment *mocks.IPaymentService) {
				mockPayment.On("GetQrMpmDynamic", mock.Anything, "12345", "", "12345").Return(nil, fmt.Errorf(""))
			},
			reqSetting:     validRequestID,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			mockPayment := mocks.NewIPaymentService(t)
			mockMerchant := mocks.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			router := chi.NewRouter()
			router.Get(
				"/payments/qr-mpm-dynamic",
				New(mockValidator, mockPayment, mockMerchant, mockRmq).QueryQrMpmDynamic,
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/qr-mpm-dynamic?"+test.query, nil)

			if test.reqSetting != nil {
				test.reqSetting(req)
			}
			if test.mockSetup != nil {
				test.mockSetup(mockPayment)
			}
			if test.setHeaders != nil {
				test.setHeaders(req)
			}
			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
		})
	}
}

func TestSNAPQueryQrMpmDynamic(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	mockIP := "1.2.3.4" // NOSONAR
	service := mocks.NewIPaymentService(t)

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)

	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/query/qr-mpm-dynamic", New(validator.New(), service, nil, nil, WithLogger(logger)).SNAPQueryQrMpmDynamic)

	merchantAuth := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
		ClientId:   uuid.NewString(),
	}
	bodyRequest := `{"originalReferenceNo": "6834f232-79af-4071-8447-50e2ea08a927","originalPartnerReferenceNo": "QR1721361975",  "serviceCode": "47"}`
	queryQrMpmDynamicReqMockType := constant.StringMockType()
	amountValue, err := decimal.NewFromString("1000.00")
	if err != nil {
		log.Fatalf("Invalid decimal value: %v", err)
	}
	tests := []struct {
		name           string
		bodyRequest    string
		merchantAuth   *merchantModel.MerchantAuthTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Merchant auth not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4015101", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR:Invalid field format originalReferenceNo",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"originalReferenceNo": false,"originalPartnerReferenceNo": "PARTNER12345",  "serviceCode": "47"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4005101", "Invalid Field Format originalReferenceNo"),
		},
		{
			name:           "ERROR:Invalid field format originalPartnerReferenceNo",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"originalReferenceNo": "QR1722498441","originalPartnerReferenceNo": false,  "serviceCode": "47"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4005101", "Invalid Field Format originalPartnerReferenceNo"),
		},
		{
			name:         "ERROR:Transaction not found",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"GetQrMpmDynamic", constant.ValueCtxMockType(), queryQrMpmDynamicReqMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, pkgErrs.New(response.HttpErrNotFound, pkgErrs.New("data not found", nil)))
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4045101", "Transaction Not Found"),
		},
		{
			name:         "ERROR:Internal error",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"GetQrMpmDynamic", constant.ValueCtxMockType(), queryQrMpmDynamicReqMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5005101", "Internal Server Error"),
		},
		{
			name:         "SUCCESS: Transaction status SUCCESS",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"GetQrMpmDynamic", constant.ValueCtxMockType(), queryQrMpmDynamicReqMockType, constant.StringMockType(), constant.StringMockType(),
				).Return(&paymentModel.PaymentResponse{
					UUID:        "47e9d0fc-05a8-43c2-a775-1191c143466c",
					ReferenceID: "QA202408020947",
					Qris: &paymentModel.PaymentQrisResponse{
						Amount: &paymentModel.Amount{
							Value:    amountValue,
							Currency: "IDR",
						},
						QrType:       "DYNAMIC",
						QrStatus:     "ACTIVE",
						QrUrl:        "https://example.com/qr.png",
						QrContent:    "00020101021126740025ID...",
						MerchantName: "Paper",
					},
					Status: "SUCCESS",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"responseCode":"2005100","responseMessage":"Successful","originalReferenceNo":"47e9d0fc-05a8-43c2-a775-1191c143466c","originalPartnerReferenceNo":"QA202408020947","serviceCode":"47","latestTransactionStatus":"00","transactionStatusDesc":"SUCCESS","amount":{"value":"1000.00","currency":"IDR"},"additionalInfo":{"qrType":"DYNAMIC","qrExpiredDate":"","qrStatus":"ACTIVE","qrContent":"00020101021126740025ID...","qrUrl":"https://example.com/qr.png","merchantName":"Paper","paymentStatus":"SUCCESS", "transactionDate":""}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/query/qr-mpm-dynamic", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/qr/qr-mpm-query")
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
