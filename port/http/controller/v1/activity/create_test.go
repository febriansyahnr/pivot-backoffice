package activityController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/activity"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {

	dataID := uuid.NewString()
	userClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	service := serviceMocks.NewIActivityService(t)
	handler := New(&config.Config{}, validatorExt.New(), service)

	router := chi.NewRouter()
	router.Post("/users/activities", handler.Create)

	tests := []struct {
		name           string
		body           string
		userClaims     *user.UserTokenClaims
		setupMocks     func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:User not found",
			setupMocks: func() {
				// Empty setup mock
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:       "ERROR:Invalid character",
			body:       `{B}`,
			userClaims: userClaims,
			setupMocks: func() {
				// Empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid character 'B' looking for beginning of object key string"}`,
		},
		{
			name:       "ERROR:Invalid data",
			body:       `{"tag":"credential-settings"}`,
			userClaims: userClaims,
			setupMocks: func() {
				// Empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[{"field":"Activity","message":"Key: 'CreateActivityReq.Activity' Error:Field validation for 'Activity' failed on the 'required' tag"}],"traceId":"","type":"API_ERROR"}, "message":"invalid validation"}`,
		},
		{
			name:       "ERROR:Some error",
			body:       `{"tag":"credential-settings","activity":"User copy client ID"}`,
			userClaims: userClaims,
			setupMocks: func() {
				service.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:       "SUCCESS",
			body:       `{"tag":"credential-settings","activity":"User copy client ID"}`,
			userClaims: userClaims,
			setupMocks: func() {
				service.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity"),
				).Return(nil).Run(func(args mock.Arguments) {
					args.Get(1).(*activityModel.Activity).ID = dataID
				})
			},
			wantStatusCode: http.StatusCreated,
			wantRespBody:   fmt.Sprintf(`{"code":"01","data":{"id":"%s"},"message":"Created"}`, dataID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/users/activities", strings.NewReader(test.body))

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
