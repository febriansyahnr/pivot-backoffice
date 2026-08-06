package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivate(t *testing.T) {
	payload := &userModel.ActivateUserRequest{
		Email:    "test@gmail.com",
		Password: "pass123",
		PIN:      "123456",
	}

	payloadRequestByte, err := json.Marshal(payload)
	assert.NoError(t, err)

	testCases := []struct {
		name            string
		requestBody     []byte
		mockSetup       func(userSvc *mockUser.IUserService)
		expectedStatus  int
		invitationToken string
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(&userModel.ValidateInvitationResponse{Email: payload.Email}, nil)

				userSvc.
					On(
						"ActivateUser",
						mock.Anything,
						mock.AnythingOfType("*user.ActivateUserRequest")).
					Return(&userModel.UserLoggedInResponse{}, nil)
			},
			expectedStatus:  200,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService) {
			},
			expectedStatus:  http.StatusBadRequest,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService) {
			},
			expectedStatus:  http.StatusBadRequest,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Service ValidateInvitationToken Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.Anything).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus:  http.StatusInternalServerError,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Service ValidateInvitationToken Email Mismatch",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.Anything).
					Return(&userModel.ValidateInvitationResponse{}, nil)
			},
			expectedStatus:  http.StatusBadRequest,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Service ActivateUser Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(&userModel.ValidateInvitationResponse{Email: payload.Email}, nil)

				userSvc.
					On(
						"ActivateUser",
						mock.Anything,
						mock.AnythingOfType("*user.ActivateUserRequest")).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus:  http.StatusInternalServerError,
			invitationToken: "token-1234",
		},
		{
			name:        "ERROR: Empty token",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService) {
			},
			expectedStatus:  http.StatusUnauthorized,
			invitationToken: "",
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

			tt.mockSetup(mockUserSvc)

			mc := New(mockValidator, mockUserSvc, nil, nil, nil, nil, cfg, secret, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/users/activate", bytes.NewBuffer(tt.requestBody))
			req.Header.Set("X-Invitation-Token", tt.invitationToken)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Activate)
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
