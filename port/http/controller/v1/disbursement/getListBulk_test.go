package disbursementController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListBulkBulkDisbursement(t *testing.T) {
	data := make([]disbursementModel.BulkDisbursement, 0)
	data = append(data, disbursementModel.BulkDisbursement{
		UUID: uuid.NewString(),
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
	validUserClaims := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name            string
		mockSetup       func(mockService *serviceMocks.IDisbursementService)
		expectedStatus  int
		funcQueryParams func() *url.Values
		userClaims      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
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
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with referenceId filter",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("referenceId", "test-uuid-123")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 500 on get list caused by service error",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IDisbursementService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IDisbursementService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup:      func(mockService *serviceMocks.IDisbursementService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
			},
			userClaims: nil,
			funcQueryParams: func() *url.Values {
				return nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIDisbursementService(t)
			tc.mockSetup(mockService)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			mc := New(cfg, nil, nil, Services{DisbursementSvc: mockService}, nil, nil)
			baseUrl := "/api/v1/disbursements/bulk/list"

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims))
			}

			req.Header.Set("Time-Zone", "Asia/Jakarta")

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.GetListBulkDisbursement)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
