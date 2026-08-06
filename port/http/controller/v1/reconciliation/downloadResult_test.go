package reconciliation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDownloadResult(t *testing.T) {

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IReconciliationService)
		setupRequest   func(req *http.Request)
		expectedStatus int
	}{
		{
			name: "Success download result",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("DownloadResult", mock.Anything, "test-uuid").Return("test-url", nil)
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"uuid": "test-uuid"}`)))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Failed: error when decode request",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				// empty
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = http.NoBody
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed: error when get download result",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("DownloadResult", mock.Anything, "test-uuid").Return("", errors.New("error"))
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"uuid": "test-uuid"}`)))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIReconciliationService(t)
			mockLogger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			tc.mockSetup(mockService)

			reconciliationController := New(mockLogger, &validator.Validate{}, mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/recon/download-result", nil)
			tc.setupRequest(req)

			chiRouterCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(reconciliationController.DownloadResult)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
