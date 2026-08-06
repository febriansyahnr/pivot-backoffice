package fds

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransaction(t *testing.T) {
	validUUID := uuid.New().String()
	invalidUUID := "invalid-uuid"

	testCases := []struct {
		desc           string
		setupMocks     func(mockSvc *mocks.IFdsService)
		id             string
		expectedStatus int
		expectResponse bool
	}{
		{
			desc: "success - valid UUID",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, (*fdscommon.UpdateRequest)(nil)).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"status": "updated",
							},
						},
					}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc:           "error - invalid UUID format",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             invalidUUID,
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:           "error - empty UUID",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             "",
			expectedStatus: http.StatusNotFound, // Chi router returns 404 for empty parameter
			expectResponse: false,
		},
		{
			desc: "error - service returns error",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, (*fdscommon.UpdateRequest)(nil)).
					Return(nil, errors.New("service error"))
			},
			id:             validUUID,
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc: "success - transaction not found (service handles gracefully)",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, (*fdscommon.UpdateRequest)(nil)).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: false,
							Code:    util.ValueToPtr("not_found"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"error": "transaction not found",
							},
						},
					}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc: "success - multiple provider responses",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, (*fdscommon.UpdateRequest)(nil)).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"status": "updated via fraudnet",
							},
						},
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("internal"),
							Message: map[string]string{
								"status": "updated via internal rules",
							},
						},
					}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc: "success - empty response from service",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, (*fdscommon.UpdateRequest)(nil)).
					Return(&[]fdscommon.UpdateResponse{}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc: "success - valid UUID format",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, "12345678-1234-1234-1234-123456789abc", (*fdscommon.UpdateRequest)(nil)).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"status": "processed",
							},
						},
					}, nil)
			},
			id:         "12345678-1234-1234-1234-123456789abc", // Valid UUID format
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Create mock service
			config := config.Config{}
			mockSvc := mocks.NewIFdsService(t)
			mockValidator := validator.New()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expectations
			tc.setupMocks(mockSvc)

			// Create controller with mock service
			controller := New(&config, mockLogger, mockValidator, mockSvc)

			// Create request with chi context for URL params
			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/internal/v1/fds/update/"+tc.id,
				nil,
			)
			assert.NoError(t, err)

			// Create test response recorder
			rr := httptest.NewRecorder()

			// Setup chi router for this test
			testRouter := chi.NewRouter()
			testRouter.Post("/internal/v1/fds/update/{id}", controller.UpdateTransaction)

			// Serve the request
			testRouter.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify response content for successful cases
			if tc.expectResponse && tc.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Body.String(), "Success") // API response format uses "Success"
				assert.Contains(t, rr.Body.String(), "data")
			}

			// Verify that all expectations were met
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction_EdgeCases(t *testing.T) {
	testCases := []struct {
		desc           string
		setupRequest   func() *http.Request
		setupMocks     func(mockSvc *mocks.IFdsService)
		expectedStatus int
	}{
		{
			desc: "success - request with various HTTP methods (though endpoint expects POST)",
			setupRequest: func() *http.Request {
				validUUID := uuid.New().String()
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPut, // Different method, but router should handle
					"/internal/v1/fds/update/"+validUUID,
					nil,
				)
				return req
			},
			setupMocks: func(mockSvc *mocks.IFdsService) {
				// No mock needed as this should fail at router level
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			desc: "error - malformed UUID patterns",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/internal/v1/fds/update/not-a-uuid-at-all",
					nil,
				)
				return req
			},
			setupMocks: func(mockSvc *mocks.IFdsService) {
				// No mock needed as this should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			desc: "error - UUID with special characters",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/internal/v1/fds/update/12345678-1234-1234-1234-12345678901g", // Invalid character 'g'
					nil,
				)
				return req
			},
			setupMocks: func(mockSvc *mocks.IFdsService) {
				// No mock needed as this should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Create mock service
			config := config.Config{}
			mockSvc := mocks.NewIFdsService(t)
			mockValidator := validator.New()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expectations
			tc.setupMocks(mockSvc)

			// Create controller with mock service
			controller := New(&config, mockLogger, mockValidator, mockSvc)

			req := tc.setupRequest()

			// Create test response recorder
			rr := httptest.NewRecorder()

			// Setup chi router for this test
			testRouter := chi.NewRouter()
			testRouter.Post("/internal/v1/fds/update/{id}", controller.UpdateTransaction)

			// Serve the request
			testRouter.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify that all expectations were met
			mockSvc.AssertExpectations(t)
		})
	}
}
