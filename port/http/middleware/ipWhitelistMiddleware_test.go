package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pdk/go/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIPWhitelistMiddleware(t *testing.T) {
	claims := merchant.MerchantAuthTokenClaims{}
	defaultIPAddress := "182.253.54.241"

	testCases := []struct {
		Name         string
		HeaderKey    string
		IPAddress    *string
		SetupRequest func(r *http.Request) *http.Request
		Claims       *merchant.MerchantAuthTokenClaims
		MockSetup    func(svc *mocks.IIPWhitelistService)
		HttpMethod   string
		ExpectedCode int
	}{
		{
			Name:         "SUCCESS",
			HeaderKey:    constant.HeaderAuthorization,
			Claims:       &claims,
			ExpectedCode: http.StatusOK,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IIPWhitelistService) {
				svc.On("ValidateIP", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:         "SUCCESS: From header",
			HeaderKey:    constant.ClientIdKey,
			ExpectedCode: http.StatusOK,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IIPWhitelistService) {
				svc.On("ValidateIP", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:         "ERROR: No Claim",
			HeaderKey:    constant.HeaderAuthorization,
			Claims:       nil,
			ExpectedCode: http.StatusUnauthorized,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IIPWhitelistService) {
			},
		},
		{
			Name:         "ERROR: Error check usecase",
			HeaderKey:    constant.HeaderAuthorization,
			Claims:       &claims,
			ExpectedCode: http.StatusInternalServerError,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IIPWhitelistService) {
				svc.On("ValidateIP", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			Name:         "SUCCESS due to empty IP",
			IPAddress:    util.ValueToPtr(""),
			HeaderKey:    constant.HeaderAuthorization,
			Claims:       &claims,
			ExpectedCode: http.StatusOK,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IIPWhitelistService) {
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			svc := mocks.NewIIPWhitelistService(t)
			tc.MockSetup(svc)
			middleware := middleware.IPWhitelistMiddleware(svc, tc.HeaderKey)
			IPAddress := defaultIPAddress
			router := chi.NewRouter()
			MountHandlers(router, middleware)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.HttpMethod, "/test", nil)

			if tc.IPAddress != nil {
				IPAddress = *tc.IPAddress
			}

			req.Header.Set(constant.HeaderXRealIP, IPAddress)
			if tc.Claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.Claims))
			}
			if tc.SetupRequest != nil {
				req = tc.SetupRequest(req)
			}

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.ExpectedCode, rec.Result().StatusCode)
		})
	}
}
