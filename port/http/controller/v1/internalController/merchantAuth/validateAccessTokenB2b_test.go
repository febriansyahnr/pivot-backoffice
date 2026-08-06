package internalMerchantAuthController

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/go-playground/validator/v10"
	merchantServiceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAccessTokenB2b(t *testing.T) {
	testCases := []struct {
		name           string
		requestSetup   func(req *http.Request)
		setup          func(mockService *merchantServiceMocks.IMerchantService)
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			requestSetup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
				req.Header.Set("X-Merchant-Id", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"ValidateAccessTokenB2b",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ERROR: Missing bearer token",
			requestSetup: func(req *http.Request) {
				req.Header.Set("X-Merchant-Id", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Validate",
			requestSetup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
				req.Header.Set("X-Merchant-Id", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"ValidateAccessTokenB2b",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			mockValidator := validator.New()
			tc.setup(mockService)
			merchantAuthController := New(mockValidator, mockService)

			baseUrl := "/api/internal/v1/access-token/b2b/validate"
			req := httptest.NewRequest(http.MethodPost, baseUrl, nil)
			tc.requestSetup(req)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.ValidateAccessTokenB2b)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)

		})
	}

}

func TestValidateSNAPAccessTokenB2b(t *testing.T) {
	testCases := []struct {
		name           string
		requestSetup   func(req *http.Request)
		setup          func(mockService *merchantServiceMocks.IMerchantService)
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			requestSetup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"ValidateAccessTokenB2b",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ERROR: Missing bearer token",
			requestSetup: func(req *http.Request) {
				req.Header.Set("X-Merchant-Id", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Validate",
			requestSetup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
			},
			setup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"ValidateAccessTokenB2b",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			mockValidator := validator.New()
			tc.setup(mockService)
			merchantAuthController := New(mockValidator, mockService)

			baseUrl := "/api/internal/v1/access-token/b2b/validate"
			req := httptest.NewRequest(http.MethodPost, baseUrl, nil)
			tc.requestSetup(req)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.ValidateSNAPAccessTokenB2b)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)

		})
	}

}
