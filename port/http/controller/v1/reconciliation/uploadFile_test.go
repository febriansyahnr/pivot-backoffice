package reconciliation

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
)

func TestUploadFile(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IReconciliationService)
		setupRequest   func(req *http.Request)
		expectedStatus int
	}{
		{
			name: "Success upload file",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				fileID := "test-file-id"
				mockService.On("UploadFile", mock.Anything, mock.Anything, "test-identifier", mock.Anything, mock.Anything).Return(&fileID, nil)
			},
			setupRequest: func(req *http.Request) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.xlsx")
				io.Copy(part, bytes.NewBuffer([]byte("test content")))
				writer.Close()

				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set(constant.XIdentifierKey, "test-identifier")
				req.Body = io.NopCloser(body)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Failed: missing x-identifier header",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {},
			setupRequest: func(req *http.Request) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.xlsx")
				io.Copy(part, bytes.NewBuffer([]byte("test content")))
				writer.Close()

				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Body = io.NopCloser(body)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failed: missing file in form",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {},
			setupRequest: func(req *http.Request) {
				req.Header.Set(constant.XIdentifierKey, "test-identifier")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed: error from service",
			mockSetup: func(mockService *serviceMocks.IReconciliationService) {
				mockService.On("UploadFile", mock.Anything, mock.Anything, "test-identifier", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			setupRequest: func(req *http.Request) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.xlsx")
				io.Copy(part, bytes.NewBuffer([]byte("test content")))
				writer.Close()

				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set(constant.XIdentifierKey, "test-identifier")
				req.Body = io.NopCloser(body)
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

			req := httptest.NewRequest(http.MethodPost, "/api/v1/reconciliation/upload", nil)
			tc.setupRequest(req)

			chiRouterCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(reconciliationController.UploadFile)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
