package internalPaymentController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestQueryQrMpmStatic(t *testing.T) {
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	testCase := []struct {
		name           string
		setupBody      func(*testing.T) []byte
		reqSetting     func(r *http.Request)
		mockSetup      func(paymentSvc *mocks.IPaymentService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
	}{
		{
			name:           "success query qr mpm static",
			wantStatusCode: http.StatusOK,
			mockSetup: func(mockPayment *mocks.IPaymentService) {
				mockPayment.On("GetQrMpmStatic", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentResponse{}, nil)
			},
			reqSetting: validRequestID,
			setHeaders: func(req *http.Request) {
				req.Header.Set("X-SubMerchantID", "12345")
			},
			setupBody: func(t *testing.T) []byte {
				body := paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
				b, err := json.Marshal(body)
				require.NoError(t, err)
				return b
			},
		},
		{
			name:           "error invalid merchant",
			wantStatusCode: http.StatusUnauthorized,
			setHeaders: func(req *http.Request) {
				req.Header.Set("X-SubMerchantID", "12345")
			},
			setupBody: func(t *testing.T) []byte {
				body := paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
				b, err := json.Marshal(body)
				require.NoError(t, err)
				return b
			},
		},
		{
			name:           "error get qr mpm static",
			wantStatusCode: http.StatusInternalServerError,
			mockSetup: func(mockPayment *mocks.IPaymentService) {
				mockPayment.On("GetQrMpmStatic", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf(""))
			},
			reqSetting: validRequestID,
			setupBody: func(t *testing.T) []byte {
				body := paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
				b, err := json.Marshal(body)
				require.NoError(t, err)
				return b
			},
		},
		{
			name:           "error request body",
			wantStatusCode: http.StatusBadRequest,
			reqSetting:     validRequestID,
			setupBody: func(t *testing.T) []byte {
				return []byte("invalid")
			},
		},
		{
			name:           "error validate request body",
			wantStatusCode: http.StatusBadRequest,
			reqSetting:     validRequestID,
			setupBody: func(t *testing.T) []byte {
				body := paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "",
				}
				b, err := json.Marshal(body)
				require.NoError(t, err)
				return b
			},
		},
	}

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			mockPayment := mocks.NewIPaymentService(t)
			mockMerchant := mocks.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			router := chi.NewRouter()
			router.Post(
				"/payments/qr-mpm-static",
				New(mockValidator, mockPayment, mockMerchant, mockRmq).QueryQrMpmStatic,
			)

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, "/payments/qr-mpm-static", bytes.NewBuffer(test.setupBody(t)))

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

func TestSNAPQueryQrMpmStatic(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	service := mocks.NewIPaymentService(t)
	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)

	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Use(openApi.IdentitySnapServiceCode)
	router.Post("/query/qr-mpm-static", New(validator.New(), service, nil, nil, WithLogger(logger)).SNAPQueryQrMpmStatic)

	merchantAuth := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
		ClientId:   uuid.NewString(),
	}
	bodyRequest := `{"partnerReferenceNo": "QR1722498441","fromDateTime": "2024-07-30T00:00:00+07:00","toDateTime": "2024-08-01T23:00:00+07:00","pageSize": 20, "pageNumber": 1}`
	queryQrMpmStaticReqMockType := mock.AnythingOfType("*paymentModel.QueryQrMpmStaticRequest")

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
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4011201", "Invalid Token (B2B)"),
		},
		{
			name:           "ERROR:Invalid field format partnerReferenceNo",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": false,"fromDateTime": false,"toDateTime": false,"pageSize": false, "pageNumber": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001201", "Invalid Field Format partnerReferenceNo"),
		},
		{
			name:           "ERROR:Invalid field format fromDateTime",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "","fromDateTime": false,"toDateTime": false,"pageSize": false, "pageNumber": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001201", "Invalid Field Format fromDateTime"),
		},
		{
			name:           "ERROR:Invalid field format toDateTime",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "","fromDateTime": "2024-07-30T00:00:00+07:00","toDateTime": false,"pageSize": false, "pageNumber": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001201", "Invalid Field Format toDateTime"),
		},
		{
			name:           "ERROR:Invalid field format pageSize",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "","fromDateTime": "2024-07-30T00:00:00+07:00","toDateTime": "2024-08-01T23:00:00+07:00","pageSize": false, "pageNumber": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001201", "Invalid Field Format pageSize"),
		},
		{
			name:           "ERROR:Invalid field format pageNumber",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "","fromDateTime": "2024-07-30T00:00:00+07:00","toDateTime": "2024-08-01T23:00:00+07:00","pageSize": 20, "pageNumber": false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001201", "Invalid Field Format pageNumber"),
		},
		{
			name:           "ERROR:Invalid mandatory field partnerReferenceNo",
			merchantAuth:   merchantAuth,
			bodyRequest:    `{"partnerReferenceNo": "","fromDateTime": "2024-07-30T00:00:00+07:00","toDateTime": "2024-08-01T23:00:00+07:00","pageSize": 20, "pageNumber": 1}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4001202", "Invalid Mandatory Field partnerReferenceNo"),
		},
		{
			name:         "ERROR:Transaction not found",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"GetQrMpmStatic", constant.ValueCtxMockType(), queryQrMpmStaticReqMockType, constant.StringMockType(),
				).Once().Return(nil, pkgErrs.New(response.HttpErrNotFound, errors.New("data not found")))
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4041201", "Transaction Not Found"),
		},
		{
			name:         "ERROR:Internal error",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				service.On(
					"GetQrMpmStatic", constant.ValueCtxMockType(), queryQrMpmStaticReqMockType, constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5001201", "Internal Server Error"),
		},
		{
			name:         "SUCCESS",
			merchantAuth: merchantAuth,
			bodyRequest:  bodyRequest,
			setupMock: func() {
				var data []paymentModel.QrStaticDetailData
				service.On(
					"GetQrMpmStatic", constant.ValueCtxMockType(), queryQrMpmStaticReqMockType, constant.StringMockType(),
				).Return(&paymentModel.PaymentResponse{
					UUID:        "47e9d0fc-05a8-43c2-a775-1191c143466c",
					ReferenceID: "QA202408020947",
					Qris: &paymentModel.PaymentQrisResponse{
						DetailData:   &data,
						QrType:       "STATIC",
						QrStatus:     "ACTIVE",
						QrUrl:        "https://sit-marketing-img.bankneo.co.id/qris/merchant/img/MZ-0ayWHXh3BnaggQPTm_CKxrKxQQS7_wqJruYi8Y2Q.png",
						QrContent:    "00020101021126740025ID.CO.BANKNEOCOMMERCE.WWW011893600490591008051102120005100009280303UBE51550025ID.CO.BANKNEOCOMMERCE.WWW0215BNC2407013497200303UBE5204078053033605802ID5910sub Harsya6006SIDRAP6105916146233052230018130955580512460800703A0163040BA8",
						MerchantName: "Paper",
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"responseCode":"2001200","responseMessage":"Successful","referenceNo":"47e9d0fc-05a8-43c2-a775-1191c143466c","partnerReferenceNo":"QA202408020947","detailData":null,"additionalInfo":{"qrType":"STATIC","qrStatus":"ACTIVE","qrExpiredDate":"","qrContent":"00020101021126740025ID.CO.BANKNEOCOMMERCE.WWW011893600490591008051102120005100009280303UBE51550025ID.CO.BANKNEOCOMMERCE.WWW0215BNC2407013497200303UBE5204078053033605802ID5910sub Harsya6006SIDRAP6105916146233052230018130955580512460800703A0163040BA8","qrUrl":"https://sit-marketing-img.bankneo.co.id/qris/merchant/img/MZ-0ayWHXh3BnaggQPTm_CKxrKxQQS7_wqJruYi8Y2Q.png","merchantName":"Paper"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/query/qr-mpm-static", strings.NewReader(test.bodyRequest))
			req.Header.Set(constant.HeaderXSnapPath, "/api/snap/v1.0/qr/transaction-history-list")
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
