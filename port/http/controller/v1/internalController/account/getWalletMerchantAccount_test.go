package internalAccountController

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetWalletMerchantAccount(t *testing.T) {

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		query          string
		setupMock      func(accountSvc *serviceMocks.IAccountService)
		setHeaders     func(req *http.Request) *http.Request
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "SUCCESS",
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"GetWalletMerchantAccount",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Return(
					&accountModel.Account{
						Name:     constant.TypeDisbursement,
						Currency: "IDR",
					},
					nil,
				)
			},
			setHeaders: func(req *http.Request) *http.Request {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())

				return req
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"00000000-0000-0000-0000-000000000000","entityId":"00000000-0000-0000-0000-000000000000","name":"DISBURSEMENT","currency":"IDR","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "ERROR: Invalid merchant id",
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
			},
			setHeaders: func(req *http.Request) *http.Request {
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
				return req
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid merchant id value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Get wallet merchant account",
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"GetWalletMerchantAccount",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			setHeaders: func(req *http.Request) *http.Request {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
				req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
				return req
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"errors","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accountSvc := serviceMocks.NewIAccountService(t)
			handler := New(accountSvc, nil)

			rec := httptest.NewRecorder()

			url := "/wallet/merchants/%v/account"
			url = fmt.Sprintf(url, uuid.NewString())

			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				req = tc.setHeaders(req)
			}

			router := chi.NewRouter()
			router.Get(url, handler.GetWalletMerchantAccount)
			tc.setupMock(accountSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, tc.wantRespBody, rec.Body.String(), "expected: %s \tgot: %s", rec.Body.String(), tc.wantRespBody)
		})

	}
}
