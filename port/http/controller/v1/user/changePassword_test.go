package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	userMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_ChangePassword(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payload := user.ChangePasswordRequest{
		NewPassword: "*****",
		OldPassword: "*****",
	}
	payloadRequestByte, err := json.Marshal(payload)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		mockSetup      func(mockService *userMocks.IUserService, mockRmq *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		userClaim      *user.UserTokenClaims
		expectedStatus int
	}{
		{
			name: "SUCCESS: Change password",
			mockSetup: func(mockService *userMocks.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&user.ChangePasswordResponse{
					Updated: true,
				}, nil)

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
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR: Invalid JSON",
			expectedStatus: http.StatusBadRequest,
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup: func(mockService *userMocks.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			userClaim: validUserClaims,
		},
		{
			name:      "ERROR: Missing required request",
			mockSetup: func(mockService *userMocks.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := user.ChangePasswordRequest{}
				invalidPayloadRequestByte, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return invalidPayloadRequestByte
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "ERROR: User not in Context",
			mockSetup: func(mockService *userMocks.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				return payloadRequestByte
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *userMocks.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("some-error"))
			},
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				return payloadRequestByte
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockService := userMocks.NewIUserService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			mockValidator := validator.New()
			tt.mockSetup(mockService, mockRmq)
			userController := New(mockValidator, mockService, nil, nil, nil, nil, nil, nil, mockRmq, nil)

			baseUrl := "/api/v1/change-password"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(userController.ChangePassword)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
