package callbackController

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetList(t *testing.T) {
	data := make([]callbackModel.Callback, 0)
	data = append(data, callbackModel.Callback{
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
		)
		expectedStatus  int
		funcQueryParams func() *url.Values
		userClaim       *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.ICallbackService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.GetListCallbackFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaim,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.ICallbackService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.GetListCallbackFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				return &queryParams
			},
			userClaim: validUserClaim,
		},
		{
			name: "ERROR: Got error 500 on get list caused by invalid callback service",
			mockSetup: func(mockService *serviceMocks.ICallbackService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.GetListCallbackFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaim,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup: func(mockService *serviceMocks.ICallbackService) {
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
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.ICallbackService) {
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

			tc.mockSetup(mockService)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			callbackController := New(cfg, validate, mockService, nil)

			baseUrl := "/api/v1/merchants"

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
			handler := http.HandlerFunc(callbackController.GetCallbackList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
