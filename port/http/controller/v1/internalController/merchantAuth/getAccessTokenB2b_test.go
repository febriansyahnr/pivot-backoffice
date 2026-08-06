package internalMerchantAuthController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	merchantServiceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAccessTokenB2b(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mockService *merchantServiceMocks.IMerchantService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
	}{
		{
			name: "SUCCESS: Get Access Token B2B",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				validToken := "valid-token"
				mockService.On(
					"GetAccessTokenB2b",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&validToken, nil)
			},

			setupBody: func(t *testing.T) []byte {
				payload := merchant.AccessTokenB2bRequest{
					GrantType: "client_credentials",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR: Invalid JSON",
			expectedStatus: http.StatusBadRequest,
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {},
		},
		{
			name:      "ERROR: Missing required request",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {},
			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"GetAccessTokenB2b",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("invalid-token"))
			},
			setupBody: func(t *testing.T) []byte {
				payload := merchant.AccessTokenB2bRequest{
					GrantType: "client_credentials",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			mockValidator := validator.New()
			tc.mockSetup(mockService)
			merchantAuthController := New(mockValidator, mockService)

			baseUrl := "/api/internal/v1/access-token/b2b"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tc.setupBody(t)))
			req.Header.Set(constant.ClientIdKey, "uuid-uuid-uuid-uuid")
			req.Header.Set(constant.ClientSecretKey, "test-secret")

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.GetAccessTokenB2b)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSNAPAccessTokenB2B(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantSvc := merchantServiceMocks.NewIMerchantService(t)

	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)
	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Post("/access-token", New(nil, merchantSvc, WithLogger(logger)).GetSNAPAccessTokenB2B)

	headers := http.Header{}
	bodyRequest := `{"grantType": "client_credentials"}`

	tests := []struct {
		name           string
		bodyReq        string
		setupReq       func(r *http.Request)
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			bodyReq:        `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4007301", "Invalid Field Format (JSON Format)"),
		},
		{
			name:           "ERROR:Invalid field format",
			bodyReq:        `{"grantType":false}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4007301", "Invalid Field Format grantType"),
		},
		{
			name:           "ERROR:Empty header X-Timestamp",
			bodyReq:        bodyRequest,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4007302", "Invalid Mandatory Field Header X-Timestamp"),
		},
		{
			name:    "ERROR:Empty header X-Client-Key",
			bodyReq: bodyRequest,
			setupReq: func(r *http.Request) {
				r.Header = headers

				headers.Set(constant.HeaderXTimestamp, "2024-08-08T13:25:45+07:00")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4017300", "Unauthorized X-Client-Key"),
		},
		{
			name:    "ERROR:Empty header X-Signature",
			bodyReq: bodyRequest,
			setupReq: func(r *http.Request) {
				r.Header = headers

				headers.Set(constant.HeaderXClientKey, "client-key")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4017300", "Unauthorized X-Signature"),
		},
		{
			name:    "ERROR:Empty grandType request body",
			bodyReq: `{}`,
			setupReq: func(r *http.Request) {
				r.Header = headers

				headers.Set(constant.HeaderXSignature, "signature")
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("4007302", "Invalid Mandatory Field grantType"),
		},
		{
			name:     "ERROR:Some error",
			bodyReq:  bodyRequest,
			setupReq: func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantSvc.On(
					"GetSNAPAccessTokenB2B", constant.ValueCtxMockType(), mock.AnythingOfType("*merchant.SNAPAccessTokenB2BReq"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrOpenApiSnapForTest("5007301", "Internal Server Error"),
		},
		{
			name:     "SUCCESS",
			bodyReq:  bodyRequest,
			setupReq: func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantSvc.On(
					"GetSNAPAccessTokenB2B", constant.ValueCtxMockType(), mock.AnythingOfType("*merchant.SNAPAccessTokenB2BReq"),
				).Return(&merchant.SNAPAccessTokenB2BResp{
					ResponseCode:    "2007300",
					ResponseMessage: "Successful",
					AccessToken:     "token", TokenType: "Bearer", ExpiresIn: "900",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"responseCode":"2007300","responseMessage":"Successful","accessToken":"token","tokenType":"Bearer","expiresIn":"900"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/access-token", strings.NewReader(test.bodyReq))

			if test.setupReq != nil {
				test.setupReq(req)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestGenerateB2BTokenSNAPSignature(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantSvc := merchantServiceMocks.NewIMerchantService(t)
	mockValidator := validator.New()

	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)
	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Post("/access-token", New(mockValidator, merchantSvc, WithLogger(logger)).GenerateB2BTokenSNAPSignature)

	headers := http.Header{}

	tests := []struct {
		name           string
		bodyReq        string
		setupReq       func(r *http.Request)
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			bodyReq:        `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid character 'A' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:    "ERROR:Empty grandType request body",
			bodyReq: `{"clientId":123}`,
			setupReq: func(r *http.Request) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"json: cannot unmarshal number into Go struct field GenerateSnapB2BTokenSignatureRequest.clientId of type string","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:    "ERROR:Some error",
			bodyReq: `{"clientId":"5ffd4643-d129-433f-85cb-cd5ebb3f17a6","timestamp":"2020-01-01T00:00:00+07:00","privateKey":"privateKey", "grantType":"grant"}`,
			setupReq: func(r *http.Request) {
			},
			setupMock: func() {
				merchantSvc.On(
					"GenOpenAPISignature", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return("", constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:     "SUCCESS",
			bodyReq:  `{"clientId":"5ffd4643-d129-433f-85cb-cd5ebb3f17a6","timestamp":"2020-01-01T00:00:00+07:00","privateKey":"privateKey", "grantType":"grant"}`,
			setupReq: func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantSvc.On(
					"GenOpenAPISignature", constant.ValueCtxMockType(), mock.Anything,
				).Return("signature", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"signature":"signature"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/access-token", strings.NewReader(test.bodyReq))

			if test.setupReq != nil {
				test.setupReq(req)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), "want %s, got %s", test.wantRespBody, rec.Body.String())
		})
	}
}

func TestValidateB2B2CTokenSNAPSignature(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantSvc := merchantServiceMocks.NewIMerchantService(t)
	mockValidator := validator.New()

	mockIP := "1.2.3.4" // NOSONAR

	mntr, err := monitor.New("testing", mockIP, "5555")
	require.NoError(t, err)
	require.NotNil(t, mntr)
	monitor.SetGlobalMonitoring(mntr)

	router := chi.NewRouter()
	router.Post("/validate", New(mockValidator, merchantSvc, WithLogger(logger)).ValidateB2B2CTokenSNAPSignature)

	headers := http.Header{}

	tests := []struct {
		name           string
		bodyReq        string
		setupReq       func(r *http.Request)
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid request body",
			bodyReq:        `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid character 'A' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:    "ERROR: Empty client id",
			bodyReq: `{"clientId":"","timestamp":"2020-01-01T00:00:00+07:00","signature":"signature"}`,
			setupReq: func(r *http.Request) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"ClientId","message":"Key: 'SNAPValidateB2b2cTokenSignatureRequest.ClientId' Error:Field validation for 'ClientId' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:    "ERROR:Some error",
			bodyReq: `{"clientId":"5ffd4643-d129-433f-85cb-cd5ebb3f17a6","timestamp":"2020-01-01T00:00:00+07:00","signature":"signature"}`,
			setupReq: func(r *http.Request) {
			},
			setupMock: func() {
				merchantSvc.On(
					"ValidateSNAPAccessTokenRequestSignature", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:     "SUCCESS",
			bodyReq:  `{"clientId":"5ffd4643-d129-433f-85cb-cd5ebb3f17a6","timestamp":"2020-01-01T00:00:00+07:00","signature":"signature"}`,
			setupReq: func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantSvc.On(
					"ValidateSNAPAccessTokenRequestSignature", constant.ValueCtxMockType(), mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":"ok"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(test.bodyReq))

			if test.setupReq != nil {
				test.setupReq(req)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), "want %s, got %s", test.wantRespBody, rec.Body.String())
		})
	}
}
