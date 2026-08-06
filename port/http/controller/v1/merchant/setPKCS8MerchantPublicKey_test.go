package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetPKCS8MerchantPublicKey(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequest := &merchantModel.MerchantPublicKeyRequest{
		PublicKey: "pub-key",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"SetMerchantPublicKey",
						mock.Anything,
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"SetMerchantPublicKey",
						mock.Anything,
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			mockSetup:      func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc, mockRmq)

			mc := New(mockMerchantSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/merchants/set-public-key", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.SetPKCS8MerchantPublicKey)
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
