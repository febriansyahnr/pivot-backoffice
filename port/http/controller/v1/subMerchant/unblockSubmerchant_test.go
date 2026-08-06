package subMerchant

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnblockSubmerchant(t *testing.T) {
	merchantID := uuid.NewString()
	testcases := []struct {
		Name           string
		Request        []byte
		MockSetup      func(svc *mockSvc.IMerchantForbiddenUseCaseService)
		WantErr        bool
		ExpectedStatus int
	}{
		{
			Name:    "ERROR: Bad Request",
			Request: []byte("{invalid JSON"),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
			},
			WantErr:        true,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "ERROR: Failed Validation",
			Request: []byte(`{
				"merchantId": "123",
				"usecase":"123"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
			},
			WantErr:        true,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "ERROR: Error block usecase",
			Request: []byte(`{
				"merchantId": "` + merchantID + `",
				"usecase":"DISBURSEMENT"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
				svc.On("UnblockUseCase", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr:        true,
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name: "SUCCESS",
			Request: []byte(`{
				"merchantId": "` + merchantID + `",
				"usecase":"DISBURSEMENT"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
				svc.On("UnblockUseCase", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr:        false,
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			ctx := context.Background()
			mockMerchantSvc := mockSvc.NewIMerchantService(t)
			mockAccountSvc := mockSvc.NewIAccountService(t)
			mockOrchestratorSvc := mockSvc.NewIOrchestratorService(t)
			mockForbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.MockSetup(mockForbiddenSvc)

			controller := New(mockMerchantSvc, mockAccountSvc, mockOrchestratorSvc, mockForbiddenSvc, mockValidator, mockRmq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/sub-merchants/unblock", bytes.NewBuffer(tt.Request))
			rr := httptest.NewRecorder()

			ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: merchantID,
			})
			chiRouterCtx := chi.NewRouteContext()
			ctx = context.WithValue(ctx, chi.RouteCtxKey, chiRouterCtx)
			req = req.WithContext(ctx)
			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.Unblock)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.ExpectedStatus, rr.Code)

		})
	}
}
