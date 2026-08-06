package callbackController_test

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rabbitMqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	callbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/callback"
)

func TestGetCallbackLogList(t *testing.T) {
	data := make([]callbackModel.CallbackLogWithMaster, 0)
	data = append(data, callbackModel.CallbackLogWithMaster{
		UUID: uuid.New(),
	})
	expectedResponse := &commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	validUserClaim := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name      string
		mockSetup func(
			mockService *serviceMocks.ICallbackService,
			rmq *rabbitMqMock.RabbitMQExt,
		)
		expectedStatus  int
		funcQueryParams func() *url.Values
		userClaim       *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackLogList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrGetListCallbackLogFilterRequestMockType(),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)

				rmq.On(
					"PublishActivity", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.PtrGetListCallbackLogFilterRequestMockType()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaim,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackLogList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrGetListCallbackLogFilterRequestMockType(),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)

				rmq.On(
					"PublishActivity", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.PtrGetListCallbackLogFilterRequestMockType()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				queryParams.Add("perPage", "2")
				queryParams.Add("startUpdatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endUpdatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "ERROR: Got error 500 on get list caused by invalid callback service",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				mockService.On(
					"GetCallbackLogList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrGetListCallbackLogFilterRequestMockType(),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)

				rmq.On(
					"PublishActivity", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.PtrGetListCallbackLogFilterRequestMockType()).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaim,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid perPage",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("perPage", "invalid perPage format")
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startUpdatedAt", "invalid startUpdatedAt format")
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endUpdatedAt", "invalid endUpdatedAt format")
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.ICallbackService, rmq *rabbitMqMock.RabbitMQExt) {
				// Empty setup
			},
			userClaim: nil,
			funcQueryParams: func() *url.Values {
				return nil
			},
			expectedStatus: http.StatusUnauthorized,
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

			baseUrl := "/api/v1/callbacks"

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(callbackCtrl.GetCallbackLogList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
