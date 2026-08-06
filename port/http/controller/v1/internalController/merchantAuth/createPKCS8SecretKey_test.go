package internalMerchantAuthController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	errors "errors"

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

func TestInternalMerchantAuthController_CreatePKCS8SecretKey(t *testing.T) {
	testCases := []struct {
		desc           string
		mockSetup      func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService)
		expectedStatus int
		setHeaders     func(req *http.Request)
	}{
		{
			desc: "SUCCESS: Create PKCS8 Secret Key",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				*ctx = context.WithValue(*ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})

				mockService.On("CreatePKCS8SecretKey", mock.Anything, mock.AnythingOfType("string")).Return(&merchant.PKCS8SecretKeyResponse{
					MerchantID:        uuid.NewString(),
					MerchantPublicKey: "publicKey",
					SnapPrivateKey:    "privateKey",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "SUCCESS: Create PKCS8 Secret Key in behalf of submerchant",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				*ctx = context.WithValue(*ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})

				mockService.On("CreatePKCS8SecretKey", mock.Anything, mock.AnythingOfType("string")).Return(&merchant.PKCS8SecretKeyResponse{
					MerchantID:        uuid.NewString(),
					MerchantPublicKey: "publicKey",
					SnapPrivateKey:    "privateKey",
				}, nil)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "ERROR: Error create PKCS8 get merchantCtx from context",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				// not used
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			desc: "ERROR: Error create PKCS8 from context resp code 500",
			mockSetup: func(ctx *context.Context, mockService *merchantServiceMocks.IMerchantService) {
				*ctx = context.WithValue(*ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})

				mockService.On("CreatePKCS8SecretKey", mock.Anything, mock.AnythingOfType("string")).Return(nil, pkgErrors.New(responseHttp.HttpErrInternal, errors.New("error get merchantID from context")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			ctx := context.Background()

			intMerchantAuthController := New(nil, mockService)

			tc.mockSetup(&ctx, mockService)

			baseUrl := "/api/internal/v1/secret-key"
			req := httptest.NewRequest(http.MethodPost, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(intMerchantAuthController.CreatePKCS8SecretKey)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
