package internal_merchant

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestGetJITApiKey(t *testing.T) {

	testCases := []struct {
		name             string
		setup            func(merchantSvc *mockSvc.IMerchantService)
		setupParam       func(chi *chi.Context)
		expectedCode     int
		expectedResponse string
	}{
		{
			name: "SUCCESS: get merchant jit api Key",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"GetOrGenerateJITApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					"api-key",
					nil,
				)
			},
			setupParam: func(chi *chi.Context) {
				chi.URLParams.Add("merchantId", uuid.NewString())
			},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"apiKey":"YXBpLWtleQ=="}}`,
		},
		{
			name: "ERROR: Invalid merchant id",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
			},
			setupParam: func(chi *chi.Context) {
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"merchant id is not valid","error":{"type":"API_ERROR","message":"merchant id is not valid","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Get merchant jit api key",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"GetOrGenerateJITApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					"",
					errors.New("error"),
				)
			},
			setupParam: func(chi *chi.Context) {
				chi.URLParams.Add("merchantId", uuid.NewString())
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantSvc := mockSvc.NewIMerchantService(t)
			tc.setup(merchantSvc)
			ctrl := New(nil, merchantSvc, nil)

			baseUrl := "/internal/v1/merchants/{merchantId}/jit-api-key"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			if tc.setupParam != nil {
				tc.setupParam(chiRouterCtx)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetJITApiKey)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedCode, httpRecorder.Code)
			assert.JSONEqf(t, tc.expectedResponse, httpRecorder.Body.String(), "response not match. Expect %v, got %v", tc.expectedResponse, httpRecorder.Body.String())
		})
	}
}
