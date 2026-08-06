package internalAccountController

import (
	"bytes"
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

func TestGetWalletCustomerAccount(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		query          string
		setupMock      func(accountSvc *serviceMocks.IAccountService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:          "SUCCESS: With submerchant",
			merchantClaim: validMerchantClaim,
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"GetWalletCustomerAccount",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*account_model.GetCustomerAccountRequest"),
				).Return(
					&accountModel.Account{
						Name:     constant.TypeDisbursement,
						Currency: "IDR",
					},
					nil,
				)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"00000000-0000-0000-0000-000000000000","entityId":"00000000-0000-0000-0000-000000000000","name":"DISBURSEMENT","currency":"IDR","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:          "SUCCESS",
			merchantClaim: validMerchantClaim,
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"GetWalletCustomerAccount",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*account_model.GetCustomerAccountRequest"),
				).Return(
					&accountModel.Account{
						Name:     constant.TypeDisbursement,
						Currency: "IDR",
					},
					nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"00000000-0000-0000-0000-000000000000","entityId":"00000000-0000-0000-0000-000000000000","name":"DISBURSEMENT","currency":"IDR","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:          "ERROR: Get wallet customer account",
			merchantClaim: validMerchantClaim,
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"GetWalletCustomerAccount",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*account_model.GetCustomerAccountRequest"),
				).Return(
					nil,
					errors.New("errors"),
				)
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

			url := "/wallet/customers/%v/account"
			url = fmt.Sprintf(url, uuid.NewString())

			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("customerId", uuid.NewString())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			router := chi.NewRouter()
			router.Get(url, handler.GetWalletCustomerAccount)
			tc.setupMock(accountSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, tc.wantRespBody, rec.Body.String(), "expected: %s \tgot: %s", rec.Body.String(), tc.wantRespBody)
		})

	}
}

func TestCreateWalletCustomerAccount(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		requestBody    []byte
		setupMock      func(accountSvc *serviceMocks.IAccountService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:          "SUCCESS: Create Customer Account",
			merchantClaim: validMerchantClaim,
			requestBody:   []byte(`{"customerId":"00000000-0000-0000-0000-000000000000"}`),
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"CreateAccount",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*account_model.NewAccountRequest"),
				).Return(
					&accountModel.AccountResponse{
						Name:     constant.TypeWallet,
						Currency: "IDR",
					},
					nil,
				)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"00000000-0000-0000-0000-000000000000","merchantId":"00000000-0000-0000-0000-000000000000","name":"WALLET","eodBalance":0,"currency":"IDR","type":"","userType":"","lastUpdateBalanceAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","current_balance_check_time":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:          "ERROR: Decode Request Body",
			merchantClaim: validMerchantClaim,
			requestBody:   []byte(`{"customerId":"00000000-0000-0000-0000-000000000000",}`),
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:          "ERROR: Validate Request Body",
			merchantClaim: validMerchantClaim,
			requestBody:   []byte(`{"customerId":"123"}`),
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"CustomerID","message":"Key: 'CreateCustomerAccountRequest.CustomerID' Error:Field validation for 'CustomerID' failed on the 'uuid' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:          "ERROR: Create Customer Account",
			merchantClaim: validMerchantClaim,
			requestBody:   []byte(`{"customerId":"00000000-0000-0000-0000-000000000000"}`),
			setupMock: func(accountSvc *serviceMocks.IAccountService) {
				accountSvc.On(
					"CreateAccount",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*account_model.NewAccountRequest"),
				).Return(
					nil,
					errors.New("errors"),
				)
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

			url := "/wallet/customers/account"

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(tc.requestBody))
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			router := chi.NewRouter()
			router.Post(url, handler.CreateWalletCustomerAccount)
			tc.setupMock(accountSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, tc.wantRespBody, rec.Body.String(), "expected: %s \tgot: %s", rec.Body.String(), tc.wantRespBody)
		})

	}
}
