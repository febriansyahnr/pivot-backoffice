package shortlink

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortLinkController_GetByCode(t *testing.T) {
	testConfig := &config.Config{
		ShortLinkRedirection: config.ShortLinkRedirectionConfig{
			InvalidURL: "https://example.com/invalid",
		},
	}

	testCases := []struct {
		name           string
		code           string
		mockParam      func(r *http.Request) *http.Request
		mockSetup      func(*mockService.IShortLinkService)
		expectedStatus int
		expectedHeader map[string]string
	}{
		{
			name: "SUCCESS: valid code redirects to destination",
			code: "ABC123",
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
				mockSvc.On(
					"GetByCode",
					mock.Anything,
					"ABC123",
				).Return(&shortLinkModel.ShortLink{
					UUID:           "test-uuid",
					Reference:      "payment-ref-123",
					Code:           "ABC123",
					DestinationURL: "https://destination.com/payment/123",
				}, nil)
			},
			expectedStatus: http.StatusFound,
			expectedHeader: map[string]string{
				"Location": "https://destination.com/payment/123",
			},
		},
		{
			name: "ERROR: Empty Code",
			code: "",
			mockParam: func(r *http.Request) *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("code", "")
				return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			},
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
			},
			expectedStatus: http.StatusSeeOther,
			expectedHeader: map[string]string{
				"Location": "https://example.com/invalid",
			},
		},
		{
			name: "ERROR: not found redirects to invalid URL",
			code: "NOTFOUND",
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
				mockSvc.On(
					"GetByCode",
					mock.Anything,
					"NOTFOUND",
				).Return(nil, errPkg.New(response.HttpErrNotFound, errors.New("link not found")))
			},
			expectedStatus: http.StatusSeeOther,
			expectedHeader: map[string]string{
				"Location": "https://example.com/invalid",
			},
		},
		{
			name: "ERROR: expired link redirects to invalid URL",
			code: "EXPIRED",
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
				mockSvc.On(
					"GetByCode",
					mock.Anything,
					"EXPIRED",
				).Return(nil, errPkg.New(response.HttpErrNotFound, errors.New("link expired")))
			},
			expectedStatus: http.StatusSeeOther,
			expectedHeader: map[string]string{
				"Location": "https://example.com/invalid",
			},
		},
		{
			name: "ERROR: internal server error redirects to invalid URL",
			code: "ERROR123",
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
				mockSvc.On(
					"GetByCode",
					mock.Anything,
					"ERROR123",
				).Return(nil, errPkg.New(response.HttpErrInternal, errors.New("database connection failed")))
			},
			expectedStatus: http.StatusSeeOther,
			expectedHeader: map[string]string{
				"Location": "https://example.com/invalid",
			},
		},
		{
			name: "SUCCESS: special characters in code",
			code: "Test-123_ABC",
			mockSetup: func(mockSvc *mockService.IShortLinkService) {
				mockSvc.On(
					"GetByCode",
					mock.Anything,
					"Test-123_ABC",
				).Return(&shortLinkModel.ShortLink{
					UUID:           "special-uuid",
					Reference:      "special-ref",
					Code:           "Test-123_ABC",
					DestinationURL: "https://special.com/page",
				}, nil)
			},
			expectedStatus: http.StatusFound,
			expectedHeader: map[string]string{
				"Location": "https://special.com/page",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockSvc := mockService.NewIShortLinkService(t)
			tc.mockSetup(mockSvc)

			// Setup controller
			controller := &ShortLinkController{
				config:       testConfig,
				shortLinkSvc: mockSvc,
			}

			// Setup router with URL parameter
			router := chi.NewRouter()
			router.Get("/s", controller.GetByCode)

			req := httptest.NewRequest("GET", "/s", nil)
			rr := httptest.NewRecorder()

			if tc.mockParam != nil {
				req = tc.mockParam(req)
			} else {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("code", tc.code)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			// Execute request
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check headers
			for key, expectedValue := range tc.expectedHeader {
				actualValue := rr.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
			}

			// Verify mocks
			mockSvc.AssertExpectations(t)
		})
	}
}
