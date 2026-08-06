package submerchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchantSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSubMerchant(t *testing.T) {

	result := &merchantModel.SubMerchantResponse{
		UUID: uuid.New().String(),
	}

	payloadRequest := &merchantModel.UpdateMerchantOpenApiRequest{
		Name:          "test",          // NOSONAR
		Description:   "test",          // NOSONAR
		Address:       "malang",        // NOSONAR
		PostCode:      "123",           // NOSONAR
		Logo:          "test",          // NOSONAR
		MerchantEmail: "test@test.com", // NOSONAR
		MerchantPhone: "test",          // NOSONAR
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, _ *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.On("UpdateSubMerchantOpenApi", mock.Anything, mock.Anything).Return(result, nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Error update sub merchant",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, _ *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.On("UpdateSubMerchantOpenApi", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			expectedStatus: 500,
		},
		{
			name:        "ERROR: Sub merchant not found",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, _ *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.On("UpdateSubMerchantOpenApi", mock.Anything, mock.Anything).Return(nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
			},
			expectedStatus: 404,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(_ *mockMerchantSvc.IMerchantService, _ *mockRabbitMq.RabbitMQExt) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			merchantSvc := mockMerchantSvc.NewIMerchantService(t)
			accountSvc := mockMerchantSvc.NewIAccountService(t)
			orchestratorSvc := mockMerchantSvc.NewIOrchestratorService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(merchantSvc, mockRmq)

			mc := New(merchantSvc, accountSvc, orchestratorSvc, mockValidator)

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPut, "/sub-merchants/{id}", bytes.NewBuffer(tt.requestBody))
			chiRouterCtx.URLParams.Add("id", result.UUID)
			ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
				MerchantId: uuid.NewString(),
			})
			ctx = context.WithValue(ctx, chi.RouteCtxKey, chiRouterCtx)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Update)
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
