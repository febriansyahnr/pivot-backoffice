package callbackController_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rabbitMqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	callbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/callback"
)

func TestResendCallbackByID(t *testing.T) {
	validUserClaim := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name      string
		mockSetup func(
			mockService *serviceMocks.ICallbackService,
			rmq *rabbitMqMock.RabbitMQExt,
		)
		expectedStatus int
		userClaim      *user.UserTokenClaims
		disbursementID string
	}{
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
			disbursementID: uuid.NewString(),
		},
		{
			name: "ERROR: Invalid disbursement ID",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			userClaim:      validUserClaim,
			expectedStatus: http.StatusBadRequest,
			disbursementID: "invalid-uuid",
		},
		{
			name: "ERROR: ResendCallback service error",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"ResendCallback",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				rmq.On(
					"PublishActivity", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaim,
			disbursementID: uuid.NewString(),
		},
		{
			name: "SUCCESS: ResendCallback",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"ResendCallback",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				rmq.On(
					"PublishActivity", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaim,
			disbursementID: uuid.NewString(),
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

			baseUrl := "/api/v1/callbacks/histories/%s/resend"

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf(baseUrl, tc.disbursementID), nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			httpRecorder := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post(
				"/api/v1/callbacks/histories/{id}/resend", callbackCtrl.ResendCallbackByID,
			)
			router.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
