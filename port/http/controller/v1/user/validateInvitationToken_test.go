package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateInvitationToken(t *testing.T) {
	payloadRequest := &userModel.ValidateInvitationRequest{
		Token: "test123",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	expectedResponse := &userModel.ValidateInvitationResponse{
		UserID:       "user-123",
		UserName:     "Test User",
		Email:        "test@mail.com",
		MerchantName: "Test Merchant",
		MerchantID:   "merchant-123",
	}

	testCase := []struct {
		name           string
		expectedStatus int
		mockSetup      func(userSvc *mockUser.IUserService)
		requestBody    []byte
	}{
		{
			name: "SUCCESS",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			requestBody:    payloadRequestByte,
		},
		{
			name: "ERROR: Bad Request - Invalid JSON",
			mockSetup: func(_ *mockUser.IUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			requestBody:    []byte("{invalid JSON"),
		},
		{
			name: "ERROR: Bad Request - Failed Validation",
			mockSetup: func(_ *mockUser.IUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			requestBody:    []byte(`{"email": "12345abcde"}`),
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.
					On(
						"ValidateInvitationToken",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			requestBody:    payloadRequestByte,
		},
	}

	for _, tt := range testCase {
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
			jwtMock := mockJWT.NewIJwt(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(mockUserSvc)

			cfg.AppConfig.PaginationPerPage = 20
			userController := New(mockValidator, mockUserSvc, nil, nil, nil, jwtMock, cfg, secret, mockRmq, nil)
			baseUrl := "/api/v1/users/validate-invitation"

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(userController.ValidateInvitationToken)
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
