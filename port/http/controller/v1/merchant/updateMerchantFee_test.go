package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchantSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantControllerUpdateMerchantFee(t *testing.T) {
	payloadRequest := &merchantModel.UpdateMerchantFeeRequest{
		MerchantID:    uuid.NewString(),
		Amount:        1000000,
		AmountType:    "AMOUNT",
		ID:            "ID",
		Percentage:    0.0,
		DeductionType: constant.MerchantFeeDeductionTypeDirect,
		TaxType:       constant.MerchantTaxTypeInclusive,
		TaxPercentage: 0.0,
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	expectedMerchantFee := &merchantModel.MerchantFee{
		UUID:       uuid.NewString(),
		MerchantID: payloadRequest.MerchantID,
		Amount:     payloadRequest.Amount,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	testCases := []struct {
		name           string
		merchantId     string
		merchantFee    *merchantModel.MerchantFee
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			merchantId:  expectedMerchantFee.UUID,
			merchantFee: expectedMerchantFee,
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"UpdateMerchantFee",
						mock.Anything,
						mock.AnythingOfType("*merchant.UpdateMerchantFeeRequest"),
						mock.Anything).
					Return(nil)

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
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			merchantId:  expectedMerchantFee.UUID,
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			merchantId:  expectedMerchantFee.UUID,
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Merchant Fee cannot be empty",
			merchantId:  "",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Service Error",
			merchantId:  expectedMerchantFee.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"UpdateMerchantFee",
						mock.Anything,
						mock.AnythingOfType("*merchant.UpdateMerchantFeeRequest"),
						mock.Anything).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
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
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, "/crm/v1/merchants/fee/{id}", bytes.NewReader(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			chiRouterCtx.URLParams.Add("id", tt.merchantId)

			// Create the handler and serve the request
			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.UpdateMerchantFee)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("Handler response body: %s", httpRecorder.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			merchantSvc.AssertExpectations(t)
		})
	}
}
