package internalMerchantAuthController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantServiceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAuthInfo(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mockService *merchantServiceMocks.IMerchantService)
		expectedStatus int
		setupContext   func(ctx context.Context) context.Context
	}{
		{
			name: "SUCCESS: Get Auth Info",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&merchant.Merchant{}, nil)
			},
			expectedStatus: http.StatusOK,
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
		},
		{
			name:           "ERROR: Error get merchantID from context",
			mockSetup:      func(mockService *merchantServiceMocks.IMerchantService) {},
			expectedStatus: http.StatusUnauthorized,
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("merchant not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
		},
		{
			name: "ERROR: Merchant not found",
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On(
					"FindMerchantByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			mockValidator := validator.New()
			tc.mockSetup(mockService)
			merchantAuthController := New(mockValidator, mockService)

			baseUrl := "/api/internal/v1/me"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(tc.setupContext(req.Context()))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.GetAuthInfo)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}

}
