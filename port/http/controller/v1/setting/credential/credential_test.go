package credential_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/credential"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	var mockTypeCredDashboardReq = mock.AnythingOfType("*credential.CredentialDashboardReq")

	service := serviceMocks.NewICredentialService(t)
	handler := New(validatorExt.New(), nil, service)

	router := chi.NewRouter()
	router.Get("/credentials", handler.Get)

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		mockSetup      func(s *serviceMocks.ICredentialService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:User not found",
			mockSetup: func(s *serviceMocks.ICredentialService) {
				// Empty mock setup
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:       "ERROR:Get credential dashboard",
			userClaims: &user.UserTokenClaims{},
			mockSetup: func(s *serviceMocks.ICredentialService) {
				s.On(
					"Get", constant.ValueCtxMockType(), mockTypeCredDashboardReq,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:       "SUCCESS",
			userClaims: &user.UserTokenClaims{},
			mockSetup: func(s *serviceMocks.ICredentialService) {
				s.On(
					"Get", constant.ValueCtxMockType(), mockTypeCredDashboardReq,
				).Return(&credential.CredentialDashboardResp{
					ClientID: "unique-client-id",
					ClientSecrets: []credential.ClientSecretSummary{{
						ID:         "unique-secret-id",
						KeyName:    "Client Secret 1",
						LastUpdate: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
					}},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00","message":"OK", "data": {"clientId": "unique-client-id","clientSecrets": [{"id": "unique-secret-id", "keyName": "Client Secret 1", "lastUpdate": "2024-06-12T00:00:00Z"}]}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/credentials", nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}
			test.mockSetup(service)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
