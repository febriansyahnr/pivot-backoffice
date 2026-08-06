package ledgerController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactions(t *testing.T) {
	claim := merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	accountId := uuid.NewString()
	referenceID := uuid.Max.String()
	trxTimestamp, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 00:00:00")

	testCases := []struct {
		Name             string
		MockSetup        func(svc *mockSvc.ILedgerService)
		SetHeaders       func(req *http.Request)
		SetUrlParam      bool
		SetupRequest     func(r *http.Request) *http.Request
		Claim            *merchant.MerchantAuthTokenClaims
		WantErr          bool
		ExpectedCode     int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{
							{
								ReferenceID:          referenceID,
								Debit:                1000,
								Credit:               0,
								Type:                 constant.TypeDisbursement,
								Channel:              constant.ChannelVirtualAccount,
								Status:               constant.StatusSuccess,
								TransactionTimestamp: trxTimestamp,
							},
						},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 1,
							TotalPages: 1,
						},
					}, nil)
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize=10"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":[{"referenceId":"ffffffff-ffff-ffff-ffff-ffffffffffff","debit":1000,"credit":0,"type":"DISBURSEMENT","channel":"VIRTUAL_ACCOUNT","status":"SUCCESS","reason":"","reasonType":"","reasonDescription":"","transactionTimestamp":"2022-01-01T00:00:00Z"}],"meta":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}}`,
		},
		{
			Name: "SUCCESS: With Submerchant ID",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{
							{
								ReferenceID:          referenceID,
								Debit:                1000,
								Credit:               0,
								Type:                 constant.TypeDisbursement,
								Channel:              constant.ChannelVirtualAccount,
								Status:               constant.StatusSuccess,
								TransactionTimestamp: trxTimestamp,
							},
						},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 1,
							TotalPages: 1,
						},
					}, nil)
			},
			SetHeaders: func(req *http.Request) {
				req.Header.Set("X-Merchant-Id", uuid.NewString())
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize=10"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":[{"referenceId":"ffffffff-ffff-ffff-ffff-ffffffffffff","debit":1000,"credit":0,"type":"DISBURSEMENT","channel":"VIRTUAL_ACCOUNT","status":"SUCCESS","reason":"","reasonType":"","reasonDescription":"","transactionTimestamp":"2022-01-01T00:00:00Z"}],"meta":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}}`,
		},
		{
			Name: "ERROR: Invalid accountID",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "&accountId=&referenceType=type1&startDate=2022-01"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"invalid accountId","error":{"type":"API_ERROR","message":"invalid accountId","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Incorrect StartDate",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&startDate=2022-01"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"parsing time \"2022-01\" as \"2006-01-02\": cannot parse \"\" as \"-\"","error":{"type":"API_ERROR","message":"parsing time \"2022-01\" as \"2006-01-02\": cannot parse \"\" as \"-\"","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Incorrect EndDate",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"parsing time \"2022\" as \"2006-01-02\": cannot parse \"\" as \"-\"","error":{"type":"API_ERROR","message":"parsing time \"2022\" as \"2006-01-02\": cannot parse \"\" as \"-\"","recommendation":""},"data":null}`,
		},
		{
			Name: "SUCCESS: Empty Page",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{},
					}, nil)
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page="
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":null,"meta":{"page":0,"perPage":0,"totalItems":0,"totalPages":0}}}`,
		},
		{
			Name: "ERROR: Incorrect Page",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=a"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"failed parse page to integer","error":{"type":"API_ERROR","message":"failed parse page to integer","recommendation":""},"data":null}`,
		},
		{
			Name: "SUCCESS: Negative Page",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{},
					}, nil)
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=-1"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":null,"meta":{"page":0,"perPage":0,"totalItems":0,"totalPages":0}}}`,
		},
		{
			Name: "SUCCESS: Empty PageSize",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{},
					}, nil)
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize="
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":null,"meta":{"page":0,"perPage":0,"totalItems":0,"totalPages":0}}}`,
		},
		{
			Name: "ERROR: Incorrect PageSize",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize=a"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"failed parse pageSize to integer","error":{"type":"API_ERROR","message":"failed parse pageSize to integer","recommendation":""},"data":null}`,
		},
		{
			Name: "SUCCESS: Negative PageSize",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*ledger_model.GetLedgerTransactionData{},
					}, nil)
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize=-9"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":{"data":null,"meta":{"page":0,"perPage":0,"totalItems":0,"totalPages":0}}}`,
		},
		{
			Name: "ERROR: Get Ledger Transactions",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("GetLedgerTransactions", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("error"))
			},
			SetupRequest: func(r *http.Request) *http.Request {
				r.URL.RawQuery = "accountId=" + accountId + "&referenceType=type1&endDate=2022-01-01&startDate=2022-01-01&page=1&pageSize=10"
				return r
			},
			Claim:            &claim,
			SetUrlParam:      true,
			WantErr:          true,
			ExpectedCode:     500,
			ExpectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			ledgerSvc := mockSvc.NewILedgerService(t)

			if tc.MockSetup != nil {
				tc.MockSetup(ledgerSvc)
			}

			baseUrl := "/api/internal/v2/ledger/" + uuid.New().String()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			if tc.SetUrlParam {
				ctx = context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx)
				req = req.WithContext(ctx)
			}

			if tc.Claim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, tc.Claim)
				req = req.WithContext(ctx)
			}

			if tc.SetHeaders != nil {
				tc.SetHeaders(req)
			}

			if tc.SetupRequest != nil {
				req = tc.SetupRequest(req)
			}

			ctrl := New(ledgerSvc)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetTransactions)
			handler.ServeHTTP(httpRecorder, req)

			if !assert.Equal(t, tc.ExpectedCode, httpRecorder.Code) {
				t.Logf("Response: %s", httpRecorder.Body.String())
			}

			assert.JSONEqf(t, tc.ExpectedResponse, httpRecorder.Body.String(), "want: %v, got: %v", tc.ExpectedResponse, httpRecorder.Body.String())

		})
	}

}
