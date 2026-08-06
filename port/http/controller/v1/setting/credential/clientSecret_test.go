package credential_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/credential"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientSecretById(t *testing.T) {
	cfg := &config.SecuritySecret{
		RespEncryptKey: "4148660c-dd9f-48ab-868a-3aa2c7c8a2a6",
	}
	userClaims := &user.UserTokenClaims{
		MerchantId: "85a1549a-c66a-44ff-b370-666e76834d1d",
		UUID:       "5cc512dd-dfe0-46ea-b23e-4d76454a1be0",
	}
	secretID := "7837f4f0-4322-4814-9e0d-bebf55903a73"
	clientSecretMockType := mock.AnythingOfType("*credential.ClientSecretReq")

	service := serviceMocks.NewICredentialService(t)
	handler := New(validatorExt.New(), cfg, service)

	router := chi.NewRouter()
	router.Get("/credentials/client-secrets/{id}", handler.GetClientSecretById)
	router.Post("/credentials/client-secrets/{id}", handler.GenerateClientSecretById)

	tests := []struct {
		name           string
		secretID       string
		userClaims     *user.UserTokenClaims
		mockSetup      func(s *serviceMocks.ICredentialService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:     "ERROR:User not found",
			secretID: "id",
			mockSetup: func(s *serviceMocks.ICredentialService) {
				// Empty mock setup
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:       "ERROR:Invalid data",
			secretID:   "id",
			userClaims: userClaims,
			mockSetup: func(s *serviceMocks.ICredentialService) {
				// Empty mock setup
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[{"field":"SecretID","message":"Key: 'ClientSecretReq.SecretID' Error:Field validation for 'SecretID' failed on the 'uuid' tag"}],"traceId":"","type":"API_ERROR"},"message":"invalid validation"}`,
		},
		{
			name:       "ERROR:Client secret by id",
			secretID:   secretID,
			userClaims: userClaims,
			mockSetup: func(s *serviceMocks.ICredentialService) {
				s.On(
					"ClientSecretById", constant.ValueCtxMockType(), clientSecretMockType,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:       "SUCCESS",
			secretID:   secretID,
			userClaims: userClaims,
			mockSetup: func(s *serviceMocks.ICredentialService) {
				s.On(
					"ClientSecretById", constant.ValueCtxMockType(), clientSecretMockType,
				).Return(&credential.ClientSecretResp{
					Secret:     "random-string-with-fix-length",
					LastUpdate: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
					Time:       123456789,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"secret":"random-string-with-fix-length","lastUpdate":"2024-06-12T00:00:00Z","time":123456789}}`,
		},
	}
	for _, test := range tests {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(fmt.Sprintf("%s[%s]", test.name, method), func(t *testing.T) {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(method, "/credentials/client-secrets/"+test.secretID, nil)

				if test.userClaims != nil {
					req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
				}
				req.Header.Set(constant.HeaderXRequestPIN, "MTIzNDU2")
				test.mockSetup(service)

				router.ServeHTTP(rec, req)
				require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())

				if rec.Result().StatusCode == http.StatusOK {
					assert.Equal(t, encryption.GenerateHMAC(cfg.RespEncryptKey, strings.TrimSpace(rec.Body.String())), rec.Header().Get(constant.HeaderXResponseSignature))
				}
			})
		}
	}
}
