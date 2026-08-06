package internalAccountController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetBalance(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		usecase        string
		query          string
		setupMock      func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid merchant auth",
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}
		`,
		},
		{
			name:          "ERROR: Invalid currency",
			merchantClaim: validMerchantClaim,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				// empty setup mock
			},
			usecase:        constant.TypeDisbursement,
			query:          "currency=IDK",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody: `{"code":"field_value_invalid","error":{"details":[{"field":"currency","message":"unsupported currency, only IDR is supported"}],"traceId":"","type":"API_ERROR"},"message":"Format Field is invalid"}
`,
		},
		{
			name:          "ERROR: GetAccountByReferenceIDAndUsecase service error",
			merchantClaim: validMerchantClaim,
			usecase:       constant.TypeDisbursement,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}
		`,
		},
		{
			name:          "ERROR: GetAccountByReferenceIDAndUsecase not found",
			merchantClaim: validMerchantClaim,
			usecase:       constant.TypeDisbursement,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody: `{"code":"data_not_found","error":{"details":[{"field":"","message":"account not found"}],"traceId":"","type":"GATEWAY_ERROR"},"message":"The requested URL does not exist"}
		`,
		},
		{
			name:          "ERROR: GetAvailableMerchantBalance service error",
			merchantClaim: validMerchantClaim,
			usecase:       constant.TypeDisbursement,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&accountModel.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				orchestratorSvc.On("GetAvailableMerchantBalance",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(0.0, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}
		`,
		},
		{
			name:          "SUCCESS: With submerchant",
			merchantClaim: validMerchantClaim,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&accountModel.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				orchestratorSvc.On("GetAvailableMerchantBalance",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Return(19000.00, nil)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"availableBalance":{"currency":"IDR","value":"19000.00"}},"message":"Success"}`,
		},
		{
			name:          "SUCCESS",
			merchantClaim: validMerchantClaim,
			usecase:       constant.TypeDisbursement,
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&accountModel.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				orchestratorSvc.On("GetAvailableMerchantBalance",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Return(10000.00, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"availableBalance":{"currency":"IDR","value":"10000.00"}},"message":"Success"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accountSvc := serviceMocks.NewIAccountService(t)
			logger := logger.NewSlogger(logger.Config{})
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			handler := New(accountSvc, orchestratorSvc)
			WithLogger(handler, logger)

			rec := httptest.NewRecorder()

			url := "/balances"
			if tc.usecase != "" {
				url += "?usecase=" + tc.usecase
			}
			if tc.query != "" {
				url += "&" + tc.query
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			router := chi.NewRouter()
			router.Get("/balances", handler.GetBalance)
			tc.setupMock(accountSvc, orchestratorSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})

	}
}
