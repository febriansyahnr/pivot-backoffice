package activityController_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	activityController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/activity"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	data := make([]activityModel.Activity, 0)
	data = append(data, activityModel.Activity{
		ID: uuid.NewString(),
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

	testCases := []struct {
		name            string
		mockSetup       func(mockService *serviceMocks.IActivityService)
		expectedStatus  int
		funcQueryParams func() *url.Values
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.IActivityService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: 200,
			funcQueryParams: func() *url.Values {
				return nil
			},
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mockService *serviceMocks.IActivityService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: 200,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.IActivityService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: 200,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				return &queryParams
			},
		},
		{
			name: "FAILED: Got error 500 on get list caused by invalid activity logs",
			mockSetup: func(mockService *serviceMocks.IActivityService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("invalid activity logs"))
			},
			expectedStatus: 500,
			funcQueryParams: func() *url.Values {
				return nil
			},
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IActivityService) {},
			expectedStatus: 400,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", "invalid format")
				return &queryParams
			},
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IActivityService) {},
			expectedStatus: 400,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endCreatedAt", "invalid format")
				return &queryParams
			},
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup:      func(mockService *serviceMocks.IActivityService) {},
			expectedStatus: 400,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIActivityService(t)
			tc.mockSetup(mockService)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			actController := activityController.New(cfg, nil, mockService)
			baseUrl := "/api/v1/activities"
			queryParams := url.Values{}
			queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
			queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantIDKey, uuid.NewString()))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(actController.GetList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
