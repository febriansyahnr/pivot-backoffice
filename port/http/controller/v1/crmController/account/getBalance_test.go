package account

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetBalance(t *testing.T) {
	merchantId := uuid.NewString()

	testCases := []struct {
		name            string
		merchantId      string
		funcQueryParams func() *url.Values
		setupMock       func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService)
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name: "ERROR: Invalid merchant id",
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				// empty setup mock
			},
			merchantId:     "invalid-uuid",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:       "ERROR: Invalid currency",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDK")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:       "SUCCESS: usecase is empty",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
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
		{
			name:       "SUCCESS: userType is empty",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				return &queryParams
			},
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
		{
			name:       "ERROR: GetAccountByReferenceIDAndUsecase service error",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
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
			name:       "ERROR: GetAccountByReferenceIDAndUsecase not found",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}
		`,
		},
		{
			name:       "ERROR: GetAvailableMerchantBalance service error",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
			setupMock: func(accountSvc *serviceMocks.IAccountService, orchestratorSvc *serviceMocks.IOrchestratorService) {
				accountSvc.On("GetAccountByReferenceIDAndUsecase",
					constant.ValueCtxMockType(), constant.UuidMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&accountModel.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)

				orchestratorSvc.On("GetAvailableMerchantBalance",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(0.0, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:       "SUCCESS",
			merchantId: merchantId,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("currency", "IDR")
				queryParams.Add("usecase", "DISBURSEMENT")
				queryParams.Add("userType", "MERCHANT")
				return &queryParams
			},
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
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			svc := New(accountSvc, orchestratorSvc)

			baseUrl := "/crm/v1/merchants/" + tc.merchantId + "/balances"

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			tc.setupMock(accountSvc, orchestratorSvc)
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)

			router := chi.NewRouteContext()
			router.URLParams.Add("merchantId", tc.merchantId)

			rec := httptest.NewRecorder()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, router))

			httpHandler := http.HandlerFunc(svc.GetBalance)
			httpHandler.ServeHTTP(rec, req)

			if rec.Body.String() != "" {
				t.Logf("response: %s", rec.Body.String())
			}

			assert.Equal(t, tc.wantStatusCode, rec.Code)
			accountSvc.AssertExpectations(t)
		})

	}
}
