package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckPassword(t *testing.T) {
	validRequest := &userModel.CheckPasswordRequest{
		Password: "dummy-password",
	}
	validRequestByte, err := json.Marshal(validRequest)
	assert.NoError(t, err)

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt)
		expectedStatus int
		userClaims     *userModel.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: validRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				userSvc.On("CheckCurrentPassword",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			expectedStatus: 200,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{}`),
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: validRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				userSvc.On("CheckCurrentPassword",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: User not in Context",
			requestBody: validRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     nil,
		},
		{
			name:        "ERROR: Too Many Requests",
			requestBody: validRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				userSvc.On("CheckCurrentPassword",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(pkgErrors.New(response.HttpErrTooManyRequest, constant.ErrRateLimiterExceedFailedAttempts))

				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusTooManyRequests,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Rate Limit Failed Validate",
			requestBody: validRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				userSvc.On("CheckCurrentPassword",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrRateLimiterFailedValidate)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(mockUserSvc, mockRmq)

			mc := New(mockValidator, mockUserSvc, nil, nil, nil, nil, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, "/users/check-password", bytes.NewBuffer(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaims))
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.CheckCurrentPassword)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUserSvc.AssertExpectations(t)
		})
	}
}
