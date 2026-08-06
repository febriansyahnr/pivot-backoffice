package callbackController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rabbitMqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	callbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/callback"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCallbackEvents(t *testing.T) {
	expectedEvents := []callbackModel.CallbackEvent{
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Event:      "PAYOUT.DONE",
			Label:      "Payout Done",
			EventGroup: "Payout",
		},
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Event:      "PAYMENT.VIRTUAL-ACCOUNT.PAID",
			Label:      "Payment Virtual Account Paid",
			EventGroup: "Payment",
		},
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Event:      "REFUND.SUCCESS",
			Label:      "Refund Success",
			EventGroup: "Refund",
		},
	}

	validUserClaim := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get all callback events",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackEvents",
					mock.AnythingOfType("*context.valueCtx"),
				).Return(expectedEvents, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaim,
		},
		{
			name: "SUCCESS: Get empty callback events list",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackEvents",
					mock.AnythingOfType("*context.valueCtx"),
				).Return([]callbackModel.CallbackEvent{}, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaim,
		},
		{
			name: "ERROR: Service returns error",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackEvents",
					mock.AnythingOfType("*context.valueCtx"),
				).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaim,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewICallbackService(t)
			validate := validator.New()
			rmq := rabbitMqMock.NewRabbitMQExt(t)

			tc.mockSetup(mockService, rmq)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			callbackCtrl := callbackController.New(cfg, validate, mockService, rmq)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/callbacks/events", nil)
			chiRouterCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(callbackCtrl.GetCallbackEvents)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
