package user

import (
	"bytes"
	"context"
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

func TestUserController_UserDetail(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequestByte := []byte("")

	testCases := []struct {
		name           string
		mockSetup      func(mockService *userMocks.IUserService, mockRmq *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		userClaim      *user.UserTokenClaims
		expectedStatus int
	}{
		{
			name: "SUCCESS: Get user detail",
			mockSetup: func(mockService *userMocks.IUserService, mockRmq *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"UserDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&user.UserResponse{
					Email: "test@email.com",
				}, nil)
			},
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *userMocks.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"UserDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("some-error"))
			},
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockService := userMocks.NewIUserService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			mockValidator := validator.New()
			tt.mockSetup(mockService, mockRmq)
			userController := New(mockValidator, mockService, nil, nil, nil, nil, nil, nil, mockRmq, nil)

			baseUrl := "/api/v1/user-detail"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(userController.UserDetail)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
