package subMerchant

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
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSubMerchant(t *testing.T) {

	response := &merchantModel.Merchant{
		UUID: uuid.New().String(),
	}

	payloadRequest := &merchantModel.UpdateMerchantRequest{
		Name:              "test",
		ShortName:         "t",
		Description:       "test",
		Website:           "https://test.com",
		Address:           "malang",
		DistrictId:        123,
		PostCode:          "123",
		Logo:              "test",
		MerchantEmail:     "test@test.com",
		MerchantPhone:     "test",
		PICName:           "test",
		PICEmail:          "test@test.com",
		PICPhone:          "test",
		PICJobTitle:       "test",
		BusinessStructure: "test",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
				merchantSvcMocks.On("UpdateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Error update sub merchant",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
				merchantSvcMocks.On("UpdateSubMerchant", mock.Anything, mock.Anything).
					Return(nil, errors.New("error"))
			},
			expectedStatus: 500,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockMerchantSvc := mockSvc.NewIMerchantService(t)
			mockAccountSvc := mockSvc.NewIAccountService(t)
			mockOrchestratorSvc := mockSvc.NewIOrchestratorService(t)
			mockForbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc, mockRmq)

			mc := New(mockMerchantSvc, mockAccountSvc, mockOrchestratorSvc, mockForbiddenSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/sub-merchants/{id}", bytes.NewBuffer(tt.requestBody))
			chiRouterCtx.URLParams.Add("id", response.UUID)
			merchantID := uuid.NewString()
			ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: merchantID,
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
			mockMerchantSvc.AssertExpectations(t)
		})
	}

}
