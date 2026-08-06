package merchant

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantController_GetNotificationConfig(t *testing.T) {
	expectedRes := &merchantModel.MerchantNotificationConfig{
		Transaction: &merchantModel.MerchantNotificationTransactionConfig{
			Active: true,
			Events: []string{"PAYMENT_IN"},
		},
	}

	testCase := []struct {
		name           string
		userClaims     *userModel.UserTokenClaims
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantService)
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: "merchant-123",
			},
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("GetNotificationConfig",
					mock.Anything,
					"merchant-123",
				).Return(expectedRes, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "ERROR: Service Error",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: "merchant-123",
			},
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("GetNotificationConfig",
					mock.Anything,
					"merchant-123",
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "ERROR: Unauthorized (Missing Claims)",
			userClaims: nil,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc)

			mc := New(mockMerchantSvc, mockValidator, mockRmq)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants/notification-config", nil)
			ctx := req.Context()
			if tt.userClaims != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, tt.userClaims)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.GetNotificationConfig)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
