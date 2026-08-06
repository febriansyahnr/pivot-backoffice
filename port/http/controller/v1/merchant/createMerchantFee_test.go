package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchantSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantControllerCreateMerchantFee(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequest := &merchantModel.NewMerchantFeeRequest{
		MerchantID:    uuid.NewString(),
		Amount:        1000000.0,
		Reference:     "DISBURSEMENT",
		AmountType:    "AMOUNT",
		Percentage:    0.0,
		DeductionType: constant.MerchantFeeDeductionTypeDirect,
		TaxType:       constant.MerchantTaxTypeInclusive,
		TaxPercentage: 0.0,
	}

	merchantFeeResponse := &merchantModel.MerchantFeeResponse{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		Amount:     1000000,
		Reference:  "DISBURSEMENT",
		AmountType: "AMOUNT",
	}

	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"CreateMerchantFee",
						mock.Anything,
						mock.AnythingOfType("*merchant.NewMerchantFeeRequest"),
					).
					Return(merchantFeeResponse, nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"CreateMerchantFee",
						mock.Anything,
						mock.AnythingOfType("*merchant.NewMerchantFeeRequest"),
					).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			merchantSvc := mockMerchantSvc.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(merchantSvc, mockRmq)

			mc := New(merchantSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/merchants/fee", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.CreateMerchantFee)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			merchantSvc.AssertExpectations(t)
		})
	}
}
