package crmRateLimiterController

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	rateLimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMRateLimiterController_Update(t *testing.T) {
	merchantID := uuid.New().String()
	configID := uuid.New().String()

	tests := []struct {
		name         string
		merchantID   string
		configID     string
		requestBody  string
		mockSetup    func(mockSvc *mockService.IRateLimiter)
		expectedCode int
		expectError  bool
	}{
		{
			name:       "SUCCESS: Update rate limit configuration",
			merchantID: merchantID,
			configID:   configID,
			requestBody: `{
				"limit": 20,
				"burst": 10,
				"order": 2,
				"time": "HOUR",
				"variable": "PATH",
				"variableValue": "/api/payment",
				"variableMatchType": "PREFIX",
				"httpMethod": "GET",
				"status": "ACTIVE",
				"description": "Updated rate limit"
			}`,
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("Update", mock.Anything, mock.MatchedBy(func(req *rateLimiterModel.UpdateRateLimitConfiguration) bool {
					return req.ID == configID && req.MerchantID == merchantID && req.Limit == 20
				})).Return(&rateLimiterModel.RateLimitConfiguration{
					UUID:       configID,
					MerchantID: merchantID,
					Limit:      20,
				}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "ERROR: Invalid merchant ID",
			merchantID:   "invalid-uuid",
			configID:     configID,
			requestBody:  `{"limit": 20}`,
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Invalid config ID",
			merchantID:   merchantID,
			configID:     "invalid-uuid",
			requestBody:  `{"limit": 20}`,
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Invalid JSON",
			merchantID:   merchantID,
			configID:     configID,
			requestBody:  `{invalid json}`,
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:       "ERROR: Validation fails",
			merchantID: merchantID,
			configID:   configID,
			requestBody: `{
				"limit": 0,
				"time": "INVALID"
			}`,
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:       "ERROR: Service update fails",
			merchantID: merchantID,
			configID:   configID,
			requestBody: `{
				"limit": 20,
				"burst": 10,
				"order": 2,
				"time": "HOUR",
				"variable": "PATH",
				"variableValue": "/api/payment",
				"variableMatchType": "PREFIX",
				"httpMethod": "GET",
				"status": "ACTIVE"
			}`,
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("Update", mock.Anything, mock.AnythingOfType("*ratelimiter.UpdateRateLimitConfiguration")).
					Return(nil, errors.New("service error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mockService.NewIRateLimiter(t)
			controller := &CRMRateLimiterController{
				svc:       mockSvc,
				validator: validator.New(),
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			req := httptest.NewRequest(http.MethodPut, "/merchants/{merchantId}/rate-limits/{id}", 
				bytes.NewBufferString(tt.requestBody))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("merchantId", tt.merchantID)
			rctx.URLParams.Add("id", tt.configID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			controller.Update(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectError {
				assert.Contains(t, strings.ToLower(w.Body.String()), "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}