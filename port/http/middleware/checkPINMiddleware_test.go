package middleware_test

import (
	"context"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCheckPINMiddleware(pt *testing.T) {
	userSvc := serviceMocks.NewIUserService(pt)
	mockRabbitmq := mocks.NewRabbitMQExt(pt)
	handler := chi.NewRouter()
	MountHandlers(handler, CheckPINMiddleware(userSvc, mockRabbitmq))

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		requestPIN     string
		mockSetup      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:           "FAILED:Empty request PIN",
			userClaims:     &user.UserTokenClaims{},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "request PIN is required"}`,
		},
		{
			name:       "FAILED:Invalid PIN",
			userClaims: &user.UserTokenClaims{},
			requestPIN: "MTIzNDU2",
			mockSetup: func() {
				userSvc.On(
					"CheckCurrentPin", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message": "some error"}`,
		},
		{
			name:       "SUCCESS",
			userClaims: &user.UserTokenClaims{},
			requestPIN: "MTIzNDU2",
			mockSetup: func() {
				userSvc.On(
					"CheckCurrentPin", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if test.mockSetup != nil {
				test.mockSetup()
			}
			if test.requestPIN != "" {
				req.Header.Add(constant.HeaderXRequestPIN, test.requestPIN)
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			handler.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}

}
