package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckMerchantMiddleware(t *testing.T) {
	merchant := &merchantModel.Merchant{
		UUID: uuid.NewString(),
	}

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		mockSetup      func(merchantSvc *serviceMocks.IMerchantService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid Merchant ID",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXMerchantId, "abced")
			},
			mockSetup:      func(merchantSvc *serviceMocks.IMerchantService) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"incorrect merchant id","error":{"type":"API_ERROR","message":"incorrect merchant id","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Find Merchant By ID",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
			},
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.
					On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).
					Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("failed validation")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"failed validation","error":{"type":"API_ERROR","message":"failed validation","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Empty Merchant",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
			},
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(nil, nil)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"incorrect merchant id","error":{"type":"API_ERROR","message":"incorrect merchant id","recommendation":""},"data":null}`,
		},
		{
			name: "SUCCESS: Check Merchant",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
			},
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(merchant, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merchantSvc := serviceMocks.NewIMerchantService(t)
			router := chi.NewRouter()
			MountHandlers(router, middleware.CheckMerchantMiddleware(merchantSvc))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			test.reqSetting(req)
			test.mockSetup(merchantSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())

		})
	}

}
