package submerchant_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/subMerchant"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResendInvitation(t *testing.T) {
	path := "/v1/sub-merchants/users/resend-invitation"

	merchantId := "aec6636d-7a02-4d93-a4c5-006b9c235068" // NOSONAR
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, nil, validator.New())

	route := chi.NewRouter()
	route.Post(path, handler.ResendInvitation)

	subMerchantId := uuid.NewString()
	merchantClaims := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}
	tests := []struct {
		name           string
		merchantId     string
		merchantClaims *merchant.MerchantAuthTokenClaims
		subMerchantId  string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Unauthorized merchant",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"credentials_invalid","message":"Unauthorized","error":{"type":"API_ERROR","details":[{"field":"","message":"Unauthorized"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Missing Sub-Merchant id",
			merchantClaims: merchantClaims,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_required","message":"missing submerchant id","error":{"type":"API_ERROR","details":[{"field":"X-SubMerchant-Id","message":"missing submerchant id"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid request body",
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Make sure  format is correct"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid email format",
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Merchant not allowed perform",
			merchantId:     "1164c181-b891-4849-bbec-3fd37f33da89", // NOSONAR
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero@example.com"}`, // NOSONAR
			setupMock: func() {
				merchantSvc.On("SubMerchantResendInvitation", c.ValueCtxMockType(), mock.Anything).Once().Return(c.ErrMerchantNotAllowedPerformAction)
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"forbidden_access","message":"Provided API Key does not have the correct permissions to perform the operation","error":{"type":"API_ERROR","details":[{"field":"","message":"Provided API Key does not have the correct permissions to perform the operation"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Email not found",
			merchantId:     "1164c181-b891-4849-bbec-3fd37f33da89", // NOSONAR
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero@example.com"}`, // NOSONAR
			setupMock: func() {
				merchantSvc.On("SubMerchantResendInvitation", c.ValueCtxMockType(), mock.Anything).Once().Return(pkgErrs.New(response.HttpErrNotFound, c.ErrUserNotFound))
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   `{"code":"resource_missing","message":"The user with ID hero@example.com cannot be found","error":{"type":"GATEWAY_ERROR","details":[{"field":"","message":"The user with ID hero@example.com cannot be found"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Email already active",
			merchantId:     "1164c181-b891-4849-bbec-3fd37f33da89", // NOSONAR
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero@example.com"}`, // NOSONAR
			setupMock: func() {
				merchantSvc.On("SubMerchantResendInvitation", c.ValueCtxMockType(), mock.Anything).Once().Return(pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrUserAlreadyActivated))
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantRespBody:   `{"code":"unprocessable_entity","message":"user already activated","error":{"type":"API_ERROR","details":[{"field":"","message":"user already activated"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Some error",
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero@example.com"}`,
			setupMock: func() {
				merchantSvc.On("SubMerchantResendInvitation", c.ValueCtxMockType(), mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:           "SUCCESS",
			merchantClaims: merchantClaims,
			subMerchantId:  subMerchantId,
			requestBody:    `{"email":"hero@example.com"}`,
			setupMock: func() {
				merchantSvc.On("SubMerchantResendInvitation", c.ValueCtxMockType(), mock.Anything).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"message":"OK"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantId == "" {
				test.merchantId = merchantId
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(test.requestBody))

			ctx := context.WithValue(req.Context(), c.CtxMerchantIDKey, test.merchantId)
			if test.merchantClaims != nil {
				ctx = context.WithValue(ctx, c.CtxMerchantInfo, test.merchantClaims)
			}
			req = req.WithContext(ctx)

			if test.subMerchantId != "" {
				req.Header.Set(c.HeaderXSubMerchantID, test.subMerchantId)
			}

			route.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
