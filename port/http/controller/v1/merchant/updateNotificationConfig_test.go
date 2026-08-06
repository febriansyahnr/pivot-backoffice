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
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantControllerUpdateNotificationConfig(t *testing.T) {
	reqBody := &merchantModel.MerchantNotificationConfig{
		Transaction: &merchantModel.MerchantNotificationTransactionConfig{
			Active: true,
			Events: []string{"PAYMENT_IN"},
			Recipient: merchantModel.MerchantNotificationRecipient{
				Email: []*merchantModel.MerchantNotificationEmailRecipient{
					{Email: "test@example.com", Type: "PRIMARY"},
				},
			},
		},
	}

	testCase := []struct {
		name           string
		userClaims     *userModel.UserTokenClaims
		body           interface{}
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantService)
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: "merchant-123",
			},
			body: reqBody,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("UpdateNotificationConfig",
					mock.Anything,
					"merchant-123",
					mock.Anything,
				).Return(reqBody, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "ERROR: Validation Error (Invalid Email)",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: "merchant-123",
			},
			body: &merchantModel.MerchantNotificationConfig{
				Transaction: &merchantModel.MerchantNotificationTransactionConfig{
					Recipient: merchantModel.MerchantNotificationRecipient{
						Email: []*merchantModel.MerchantNotificationEmailRecipient{
							{Email: "invalid-email", Type: "PRIMARY"},
						},
					},
				},
			},
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: Service Error",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: "merchant-123",
			},
			body: reqBody,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService) {
				merchantSvcMocks.On("UpdateNotificationConfig",
					mock.Anything,
					"merchant-123",
					mock.Anything,
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "ERROR: Unauthorized (Missing Claims)",
			userClaims: nil,
			body:       reqBody,
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

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/merchants/notification-config", &buf)
			ctx := req.Context()
			if tt.userClaims != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, tt.userClaims)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.UpdateNotificationConfig)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
