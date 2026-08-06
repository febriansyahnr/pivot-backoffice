package openApi_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	jwtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"

	"github.com/go-chi/chi/v5"
	jwtModel "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	logger.On("Warn", c.ValueCtxMockType(), c.StringMockType(), c.ZapFieldMockType()).Return()
	logger.On("Error", c.ValueCtxMockType(), c.StringMockType(), c.ZapFieldMockType()).Return()

	jwt := jwtMock.NewIJwt(t)
	merchantService := serviceMocks.NewIMerchantService(t)

	headers := http.Header{}
	headers.Set(c.HeaderXSnapPath, "/api/snap/v1.0/qr/qr-mpm-generate")

	authMiddleware := NewSnapAuthMiddleware(logger, jwt, merchantService)

	router := chi.NewRouter()
	router.Use(IdentitySnapServiceCode, authMiddleware.Authorize)
	router.Post("/open-api/snap/v1/qris/generate-mpm", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(c.HeaderContentType, c.MIMEApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"OK"}`))
	})

	wrapErrRespBody := func(code, msg string) string {
		return fmt.Sprintf(`{"responseCode":"%s","responseMessage":"%s"}`, code, msg)
	}
	merchantId := uuid.NewString()
	subMerchantId := uuid.NewString()
	requestBody := `{"originalPartnerReferenceNo": "QR1721361975","serviceCode": "47"}`

	tests := []struct {
		name           string
		requestBody    string
		setupReq       func(r *http.Request)
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Empty Header Authorization",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4017301", "Invalid Token (B2B)"),
		},
		{
			name: "ERROR:Empty Header X-Timestamp",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderAuthorization, "token")

				r.Header = headers
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrRespBody("4004702", "Invalid Mandatory Field Header X-Timestamp"),
		},
		{
			name: "ERROR:Empty Header X-Signature",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXTimestamp, "2024-08-07T11:21:00+07:00")

				r.Header = headers
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014700", "Unauthorized X-Signature"),
		},
		{
			name: "ERROR:Empty Header X-Partner-Id",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXSignature, "+aXC7Y0iJsFlEtHZIGe/qsruQLv+WCBVCyAvF/BHDjK+2+OYGqwXmY6c2sz7qcKpa8OHr5/GJei/MXncothKtg==")

				r.Header = headers
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014700", "Unauthorized X-Partner-Id"),
		},
		{
			name: "ERROR:Empty Header X-External-Id",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXPartnerId, "partner-id")

				r.Header = headers
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrRespBody("4004702", "Invalid Mandatory Field Header X-External-Id"),
		},
		{
			name: "ERROR:Empty Header Channel-Id",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXExternalId, "external-id")

				r.Header = headers
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014700", "Unauthorized Channel-Id"),
		},
		{
			name: "ERROR:Channel-Id Not Registered",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXChannelId, "channel-id")

				r.Header = headers
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014700", "Unauthorized Channel-Id"),
		},
		{
			name: "ERROR:Invalid Token Format",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderXChannelId, "HARSYA")

				r.Header = headers
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name: "ERROR:Verify Merchant Token",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderAuthorization, "Bearer token-xxxx")

				r.Header = headers
			},
			setupMock: func() {
				jwt.On("VerifyMerchantToken", c.ValueCtxMockType(), "token-xxxx").Once().Return(nil, c.ErrInvalidToken)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name: "ERROR:Token Has Expired",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderAuthorization, "Bearer token-expired")

				r.Header = headers
			},
			setupMock: func() {
				jwt.On(
					"VerifyMerchantToken", c.ValueCtxMockType(), "token-expired",
				).Once().Return(&merchant.MerchantAuthTokenClaims{
					RegisteredClaims: jwtModel.RegisteredClaims{
						ExpiresAt: &jwtModel.NumericDate{Time: time.Now().UTC().Add(-1 * time.Minute)},
					},
				}, nil)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name: "ERROR:Find Merchant By Id",
			setupReq: func(r *http.Request) {
				headers.Set(c.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYWNrZW5kLXBvcnRhbCIsInN1YiI6IjNmYzk2ZGU4LWY2NWUtNGIxNi05MGExLWUyYTAwZDFiYWUyOSIsImV4cCI6MTcyMzAwNzIwOSwiY2xpZW50SWQiOiIzZmM5NmRlOC1mNjVlLTRiMTYtOTBhMS1lMmEwMGQxYmFlMjkiLCJtZXJjaGFudElkIjoiM2ZjOTZkZTgtZjY1ZS00YjE2LTkwYTEtZTJhMDBkMWJhZTI5In0.xWHV2r-svi30zz8fbq08omvgweWyykSAUDZObcQ_fCQ")

				r.Header = headers
			},
			setupMock: func() {
				jwt.On(
					"VerifyMerchantToken", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.MerchantAuthTokenClaims{
					MerchantId: merchantId,
					RegisteredClaims: jwtModel.RegisteredClaims{
						ExpiresAt: &jwtModel.NumericDate{Time: time.Now().UTC().Add(time.Minute)},
					},
				}, nil)

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrRespBody("5004701", "Internal Server Error"),
		},
		{
			name:     "ERROR:Merchant Id Not Found",
			setupReq: func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantService.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Once().Return(nil, nil)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name: "ERROR:Sub-Merchant Id",
			setupReq: func(r *http.Request) {
				r.Header = headers
				r.Header.Set(c.HeaderXSubMerchantID, subMerchantId)
			},
			setupMock: func() {
				merchantService.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Times(1).Return(&merchant.Merchant{}, nil)
				merchantService.On("FindMerchantByID", c.ValueCtxMockType(), subMerchantId).Times(1).Return(nil, c.ErrSomeErrorForUnitTest)

			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrRespBody("5004701", "Internal Server Error"),
		},
		{
			name: "ERROR:Sub-Merchant Invalid Parent Id",
			setupReq: func(r *http.Request) {
				r.Header = headers
				r.Header.Set(c.HeaderXSubMerchantID, subMerchantId)
			},
			setupMock: func() {
				merchantService.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Times(1).Return(&merchant.Merchant{UUID: merchantId}, nil)
				merchantService.On("FindMerchantByID", c.ValueCtxMockType(), subMerchantId).Times(1).Return(&merchant.Merchant{UUID: subMerchantId}, nil)

			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name:        "ERROR:Invalid Request Body Format",
			requestBody: `A`,
			setupReq: func(r *http.Request) {
				r.Header = headers
				r.Header.Set(c.HeaderXSubMerchantID, "")
			},
			setupMock: func() {

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.Merchant{UUID: merchantId}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrRespBody("4004701", "Invalid Field Format (Request Body Format)"),
		},
		{
			name:        "ERROR:Invalid Get PKCS8 Secret",
			requestBody: requestBody,
			setupReq:    func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantService.On(
					"GetPKCS8SecretKey", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, errors.New("ERROR_NOT_FOUND | secret key not found"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrRespBody("4014701", "Invalid Token (B2B)"),
		},
		{
			name:        "SUCCESS",
			requestBody: requestBody,
			setupReq:    func(r *http.Request) { r.Header = headers },
			setupMock: func() {
				merchantService.On(
					"GetPKCS8SecretKey", c.ValueCtxMockType(), merchantId,
				).Return(
					&merchant.PKCS8SecretKeyResponse{
						Data: "eyJtZXJjaGFudElkIjoiM2ZjOTZkZTgtZjY1ZS00YjE2LTkwYTEtZTJhMDBkMWJhZTI5IiwibWVyY2hhbnRTZWNyZXQiOiJ4RkRzT2hlUldzY1ZwYVBMU3NERWYxYUo0U2cyeDN3UnVXNE56Rm1RIn0=",
					}, nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/open-api/snap/v1/qris/generate-mpm", strings.NewReader(test.requestBody))

			if test.setupReq != nil {
				test.setupReq(req)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
