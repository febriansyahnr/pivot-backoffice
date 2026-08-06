package user_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/user"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func WrapErrResp(code int, errType, msg string) string {
	return fmt.Sprintf(`{"code":"%d","data":null,"error":{"details":[],"traceId":"","type":"%s"},"message":"%s"}`, code, errType, msg)
}

func TestGetInvitationURL(t *testing.T) {
	userService := serviceMock.NewIUserService(t)

	handler := New(nil, userService, nil, nil, nil, nil, nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/invitation/{encoded_email}", handler.GetInvitationURL)

	encodedEmail := "d2lkeWEuYmFndXMrdGVzdEBoYXJzeWEuY29t"

	tests := []struct {
		name           string
		encodedEmail   string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid param",
			encodedEmail:   "email@example.id",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   WrapErrResp(40, response.ErrTypeAPI, "invalid format"),
		},
		{
			name:         "ERROR:Some error on service",
			encodedEmail: encodedEmail,
			setupMock: func() {
				userService.On(
					"GetInvitationURL", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return("", constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   WrapErrResp(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:         "SUCCESS",
			encodedEmail: encodedEmail,
			setupMock: func() {
				userService.On(
					"GetInvitationURL", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return("http://localhost/invitation?token=123456789", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"url":"http://localhost/invitation?token=123456789"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/invitation/"+test.encodedEmail, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
