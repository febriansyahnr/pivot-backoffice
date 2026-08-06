package internalMerchantAuthController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantServiceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternalMerchantAuthController_GetPKCS8SecretKey(t *testing.T) {
	testCases := []struct {
		desc           string
		mockSetup      func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService)
		expectedStatus int
	}{
		{
			desc: "SUCCESS: Get PKCS8 Secret Key",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				*ctx = context.WithValue(*ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})

				mockService.On("GetPKCS8SecretKey", mock.Anything, mock.AnythingOfType("string")).Return(&merchant.PKCS8SecretKeyResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "ERROR: Error get merchantID from context resp code 500",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				*ctx = context.WithValue(*ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})

				mockService.On("GetPKCS8SecretKey", mock.Anything, mock.AnythingOfType("string")).Return(&merchant.PKCS8SecretKeyResponse{}, pkgErrors.New(responseHttp.HttpErrInternal, errors.New("error get merchantID from context")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			ctx := context.Background()

			merchantAuthController := New(nil, mockService)

			tc.mockSetup(&ctx, mockService)

			baseUrl := "/api/internal/v1/secret-key"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			req.Header.Add("X-CLIENT-KEY", "CLIENT-KEY")

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.GetPKCS8SecretKey)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
