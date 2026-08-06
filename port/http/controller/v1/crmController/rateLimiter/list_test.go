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
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	rateLimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMRateLimiterController_GetList(t *testing.T) {
	merchantID := uuid.New().String()

	tests := []struct {
		name         string
		merchantID   string
		queryParams  string
		mockSetup    func(mockSvc *mockService.IRateLimiter)
		expectedCode int
		expectError  bool
	}{
		{
			name:        "SUCCESS: Get list with default pagination",
			merchantID:  merchantID,
			queryParams: "",
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *rateLimiterModel.MerchantRateLimitRequest) bool {
					return req.MerchantID == merchantID && req.Page == 1 && req.PageSize == 20
				})).Return(&commonModel.PaginationResponse{
					Data: []*rateLimiterModel.RateLimitConfiguration{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0},
				}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:        "SUCCESS: Get list with custom pagination",
			merchantID:  merchantID,
			queryParams: "?page=2&perPage=5",
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *rateLimiterModel.MerchantRateLimitRequest) bool {
					return req.MerchantID == merchantID && req.Page == 2 && req.PageSize == 5
				})).Return(&commonModel.PaginationResponse{
					Data: []*rateLimiterModel.RateLimitConfiguration{},
					Meta: commonModel.Meta{Page: 2, PerPage: 5, TotalItems: 0},
				}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:        "SUCCESS: Get list with status filter",
			merchantID:  merchantID,
			queryParams: "?status=ACTIVE",
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *rateLimiterModel.MerchantRateLimitRequest) bool {
					return req.MerchantID == merchantID && req.Status == "ACTIVE"
				})).Return(&commonModel.PaginationResponse{
					Data: []*rateLimiterModel.RateLimitConfiguration{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0},
				}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "ERROR: Invalid merchant ID",
			merchantID:   "invalid-uuid",
			queryParams:  "",
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Invalid page parameter",
			merchantID:   merchantID,
			queryParams:  "?page=invalid",
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Invalid perPage parameter",
			merchantID:   merchantID,
			queryParams:  "?perPage=invalid",
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "ERROR: Zero page parameter",
			merchantID:   merchantID,
			queryParams:  "?page=0",
			mockSetup:    func(mockSvc *mockService.IRateLimiter) {},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:        "ERROR: Service list fails",
			merchantID:  merchantID,
			queryParams: "",
			mockSetup: func(mockSvc *mockService.IRateLimiter) {
				mockSvc.On("List", mock.Anything, mock.AnythingOfType("*ratelimiter.MerchantRateLimitRequest")).
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

			req := httptest.NewRequest(http.MethodGet, "/merchants/{merchantId}/rate-limits"+tt.queryParams, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("merchantId", tt.merchantID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			controller.GetList(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectError {
				assert.Contains(t, strings.ToLower(w.Body.String()), "error")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}