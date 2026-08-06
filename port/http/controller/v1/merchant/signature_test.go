package merchant_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenOpenAPISignature(t *testing.T) {
	service := serviceMocks.NewIMerchantService(t)

	router := chi.NewRouter()
	router.Post("/signature", New(service, validatorExt.New(), nil).GenOpenAPISignature)

	userClaim := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}
	timestamp := "2024-08-05T16:00:00+07:00"

	tests := []struct {
		name           string
		request        string
		timestamp      string
		userClaim      *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Invalid body format",
			request:        `Z`,
			userClaim:      userClaim,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, "invalid character 'Z' looking for beginning of value"),
		},
		{
			name:           "ERROR:PrivateKey is empty",
			timestamp:      timestamp,
			request:        `{"privateKey":""}`,
			userClaim:      userClaim,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"PrivateKey","message":"Key: 'GenSignatureReq.PrivateKey' Error:Field validation for 'PrivateKey' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR:Some error",
			timestamp: timestamp,
			request:   `{"privateKey":"secret-key"}`,
			userClaim: userClaim,
			setupMock: func() {
				service.On(
					"GenOpenAPISignature", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.GenSignatureReq"),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS",
			timestamp: timestamp,
			request:   `{"privateKey":"secret-key"}`,
			userClaim: userClaim,
			setupMock: func() {
				service.On(
					"GenOpenAPISignature", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.GenSignatureReq"),
				).Return("secret", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"signature":"secret"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/signature", strings.NewReader(test.request))

			req.Header.Add(c.HeaderXTimestamp, test.timestamp)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaim))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
