package merchant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantController_FindByMerchantID(t *testing.T) {
	now := time.Now()

	expectedMerchant := merchantModel.Merchant{
		UUID:      uuid.New().String(),
		Name:      "Testing",
		Logo:      "https://google.co.id/",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCase := []struct {
		name           string
		merchantId     string
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantService)
		expectedStatus int
	}{
		{
			name:       "SUCCESS",
			merchantId: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("FindMerchantByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&expectedMerchant, nil)
			},
			expectedStatus: 200,
		},
		{
			name:       "ERROR: Service Error",
			merchantId: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("FindMerchantByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "ERROR: Merchant not found",
			merchantId: "",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "ERROR: Merchant not found",
			merchantId: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("FindMerchantByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc)

			mc := New(mockMerchantSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/merchants/%s", tt.merchantId), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.merchantId)

			rr := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.FindByMerchantID)
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
