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

func TestUpdateReconDetail(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IReconciliationService)
		setupRequest   func(req *http.Request)
		expectedStatus int
	}{
		{
			name: "Success update recon detail",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("UpdateReconDetail", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", mock.AnythingOfType("*reconciliation.ReconDetail")).Return(nil)
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000","status":"true","reason":"test reason"}`)))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Failed: invalid request body",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				// empty
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer([]byte(`invalid json`)))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed: invalid UUID",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				// empty
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"id":"invalid-uuid","status":"APPROVED","reason":"test reason"}`)))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed: invalid status",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				// empty
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000","status":"INVALID","reason":"test reason"}`)))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed: service error",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("UpdateReconDetail", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", mock.AnythingOfType("*reconciliation.ReconDetail")).Return(errors.New("service error"))
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(bytes.NewBuffer(json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000","status":"true","reason":"test reason"}`)))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIReconciliationService(t)
			mockLogger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			tc.mockSetup(mockService)

			reconciliationController := New(mockLogger, &validator.Validate{}, mockService)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/reconciliation/detail", nil)
			tc.setupRequest(req)

			chiRouterCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(reconciliationController.UpdateReconDetail)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
