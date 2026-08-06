package apiLog_test

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/apiLog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetByID(t *testing.T) {
	mockService := serviceMock.NewIInboundService(t)
	h := New(mockService)

	router := chi.NewRouter()
	router.Get("/api-logs/{id}", h.GetByID)

	tests := []struct {
		name           string
		id             string
		userClaims     *user.UserTokenClaims
		setupReq       func(r *http.Request)
		setupMocks     func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "Failure - Invalid UUID",
			id:             "invalid-uuid",
			userClaims:     &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq:       func(r *http.Request) {},
			setupMocks:     func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid request payload"}`,
		},
		{
			name:       "Failure - Service Error",
			id:         "123e4567-e89b-12d3-a456-426614174000",
			userClaims: &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq:   func(r *http.Request) {},
			setupMocks: func() {
				mockService.On("GetByID", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(nil, errors.New("service error")).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"service error"}`,
		},
		{
			name:       "Success - Valid Request",
			id:         "123e4567-e89b-12d3-a456-426614174000",
			userClaims: &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-token")
			},
			setupMocks: func() {
				mockService.On("GetByID", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(&inboundModel.InboundResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"body":{},"createdAt":"0001-01-01T00:00:00Z","feature":"","headers":{},"id":"","ip":"","metadata":{},"method":"","originId":"","referenceId":"","responseBody":{},"responseTimeMs":0,"snapCompatibility":false,"statusCode":0,"traceId":"","updatedAt":"0001-01-01T00:00:00Z","url":""},"message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api-logs/"+tt.id, nil)
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tt.wantRespBody, rec.Body.String())
		})
	}
}

func TestGetSnapVersionByID(t *testing.T) {
	mockService := serviceMock.NewIInboundService(t)
	h := New(mockService)

	router := chi.NewRouter()
	router.Get("/api-logs/{id}/snap", h.GetSnapVersionByID)

	tests := []struct {
		name           string
		id             string
		userClaims     *user.UserTokenClaims
		setupReq       func(r *http.Request)
		setupMocks     func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "Failure - Invalid UUID",
			id:             "invalid-uuid",
			userClaims:     &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq:       func(r *http.Request) {},
			setupMocks:     func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid request payload"}`,
		},
		{
			name:       "Failure - Service Error",
			id:         "123e4567-e89b-12d3-a456-426614174000",
			userClaims: &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq:   func(r *http.Request) {},
			setupMocks: func() {
				mockService.On("GetSnapVersionByID", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(nil, errors.New("service error")).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"service error"}`,
		},
		{
			name:       "Success - Valid Request",
			id:         "123e4567-e89b-12d3-a456-426614174000",
			userClaims: &user.UserTokenClaims{MerchantId: "merchant-123"},
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-token")
			},
			setupMocks: func() {
				mockService.On("GetSnapVersionByID", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(&inboundModel.InboundSnapVersionResponse{
					SnapCompatibility: true,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"body":{},"createdAt":"0001-01-01T00:00:00Z","feature":"","headers":{},"id":"","ip":"","metadata":{},"method":"","originId":"","referenceId":"","responseBody":{},"responseTimeMs":0,"snapCompatibility":true,"statusCode":0,"traceId":"","updatedAt":"0001-01-01T00:00:00Z","url":""},"message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api-logs/"+tt.id+"/snap", nil)
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tt.wantRespBody, rec.Body.String())
		})
	}
}
