package callbackController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestCallbackRegisterCallback(t *testing.T) {
	merchantID := uuid.New()
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: merchantID.String(),
		Role:       constant.RoleAdmin,
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		userClaim      *user.UserTokenClaims
		expectedStatus int
	}{
		{
			name: "SUCCESS: Register Callback",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"RegisterCallback",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.RegisterCallbackRequest"),
				).Return(&callbackModel.RegisterCallbackResponse{
					CallbackMasterID: uuid.New(),
					CallbackName:     "Virtual Account",
					CallbackID:       uuid.New(),
					URL:              "https://localhost/v1/payment-callback",
					Description:      "API",
				}, nil)

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
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{
					MerchantID:  merchantID,
					Name:        "VA - Callback API",
					URL:         "https://localhost/v1/payment-callback",
					Description: "API",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "ERROR: Merchant id is invalid",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			userClaim: &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: "123e4567-e89b-12d3-a456-42661417400g",
				Role:       constant.RoleAdmin,
			},
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{
					MerchantID:  merchantID,
					Name:        "VA - Callback API",
					URL:         "https://localhost/v1/payment-callback",
					Description: "API",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR: Invalid JSON",
			expectedStatus: http.StatusBadRequest,
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			userClaim: validUserClaims,
		},
		{
			name:      "ERROR: Missing required request",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{
					MerchantID:  merchantID,
					Name:        "VA - Callback API",
					URL:         "https://localhost/v1/payment-callback",
					Description: "API",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "ERROR: Merchant ID not uuid after all",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{
					MerchantID:  merchantID,
					Name:        "VA - Callback API",
					URL:         "https://localhost/v1/payment-callback",
					Description: "API",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			userClaim: &user.UserTokenClaims{
				UUID:       validUserClaims.UUID,
				MerchantId: "hayoo mau gua hack",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"RegisterCallback",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.RegisterCallbackRequest"),
				).Return(nil, errors.New("some-error"))
			},
			userClaim: validUserClaims,
			setupBody: func(t *testing.T) []byte {
				payload := callbackModel.RegisterCallbackRequest{
					MerchantID:  merchantID,
					Name:        "VA - Callback API",
					URL:         "https://localhost/v1/payment-callback",
					Description: "API",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockService := serviceMocks.NewICallbackService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockService, mockRmq)
			cfg := &config.Config{}
			callbackController := New(cfg, mockValidator, mockService, mockRmq)

			baseUrl := "/api/v1/callbacks"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(callbackController.RegisterCallback)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
