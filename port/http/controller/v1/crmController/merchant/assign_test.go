package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMerchant_Assign(t *testing.T) {
	userId := uuid.NewString()

	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequest := &merchantModel.MerchantAssignRequest{
		UserID:     userId,
		MerchantID: uuid.NewString(),
	}

	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Return(&user.User{
						UUID: userId,
					}, nil)

				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Return(nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Conflict - User already assigned to merchant",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Return(&user.User{
						UUID:       userId,
						MerchantId: uuid.NewString(),
					}, nil)
			},
			expectedStatus: http.StatusConflict,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(mockUserSvc, mockRmq)

			mc := New(mockMerchantSvc, mockUserSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/merchants/assign", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Assign)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
