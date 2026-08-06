package reconciliation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	data := make([]reconciliation.Reconciliation, 0)
	expectedResponse := &commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 0,
			TotalPages: 0,
		},
	}

	testCases := []struct {
		name            string
		mockSetup       func(mockService *serviceMocks.IReconciliationService)
		expectedStatus  int
		funcQueryParams func() *url.Values
	}{
		{
			name: "Success get list recon",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("ListRecon", mock.Anything, mock.AnythingOfType("*reconciliation.ReconciliationFilterRequest")).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "1")
				return &queryParams
			},
		},
		{
			name: "Failed: errore parse page",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				// empty
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "a")
				return &queryParams
			},
		},
		{
			name: "Failed: error when get list recon",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("ListRecon", mock.Anything, mock.AnythingOfType("*reconciliation.ReconciliationFilterRequest")).Return(expectedResponse, errors.New("error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "1")
				return &queryParams
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIReconciliationService(t)
			mockLogger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			tc.mockSetup(mockService)

			reconciliationController := New(mockLogger, &validator.Validate{}, mockService)

			baseUrl := "/api/v1/reconciliation"
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(reconciliationController.GetList)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
