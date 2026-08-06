package crmRateLimiterController

import (
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

func TestCRMRateLimiterController_Detail(t *testing.T) {
	merchantID := uuid.New().String()
	configID := uuid.New().String()

	tests := []struct {
		name         string
		merchantID   string
		configID     string
		mockSetup    func(mockSvc *mockService.IRateLimiter)
		expectedCode int
		expectError  bool
	}{
		{
			name:       "SUCCESS: Get rate limit configuration detail",
			merchantID: merchantID,
			configID:   configID,
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("Detail", mock.Anything, merchantID, configID).Return(&rateLimiterModel.RateLimitConfiguration{
					UUID:       configID,
					MerchantID: merchantID,
					Limit:      10,
				}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "ERROR: Invalid merchant ID",
			merchantID:   "invalid-uuid",
			configID:     configID,
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Invalid config ID",
			merchantID:   merchantID,
			configID:     "invalid-uuid",
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:       "ERROR: Configuration not found",
			merchantID: merchantID,
			configID:   configID,
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("Detail", mock.Anything, merchantID, configID).
					Return(nil, errors.New("not found")).Once()
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

			req := httptest.NewRequest(http.MethodGet, "/merchants/{merchantId}/rate-limits/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("merchantId", tt.merchantID)
			rctx.URLParams.Add("id", tt.configID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			controller.Detail(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectError {
				assert.Contains(t, strings.ToLower(w.Body.String()), "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}