package walletTransaction_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/wallet/transaction"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func generateValidDateQuery() string {
	now := time.Now()
	validStartDate := now.AddDate(0, 0, -7)         // 7 days ago (definitely within 6 months)
	validEndDate := validStartDate.AddDate(0, 0, 2) // 2 days later (5 days ago)
	return fmt.Sprintf("startDate=%s&endDate=%s",
		url.QueryEscape(validStartDate.Format("2006-01-02T15:04:05-07:00")),
		url.QueryEscape(validEndDate.Format("2006-01-02T15:04:05-07:00")))
}

func TestGetMerchantTransactionHistoryList(t *testing.T) {
	validDateQuery := generateValidDateQuery()

	vld := validatorExt.New()
	service := serviceMock.NewIWalletTransactionService(t)

	handler := New(vld, service)

	route := chi.NewRouter()
	route.Get("/transactions", handler.GetMerchantTransactionHistoryList)

	userClaims := &user.UserTokenClaims{
		MerchantId: "431fc9ae-4026-4190-a2bd-83c1ccf7cc6a", // NOSONAR
	}

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		query          string
		timezone       string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Invalid mandatory field",
			userClaims:     userClaims,
			query:          "startDate=2025-02-27T00:00:00%2b07:00&endDate=2025-02-03T23:59:59%2b07:00",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"EndDateReq","message":"Key: 'MerchantTransactionHistoryListReq.EndDateReq' Error:Field validation for 'EndDateReq' failed on the 'gtecsfield' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Time range exceeds the limit",
			userClaims:     userClaims,
			query:          "startDate=2025-01-01T00:00:00%2b07:00&endDate=2025-03-03T23:59:59%2b07:00", // NOSONAR
			timezone:       constant.TimeLoc,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"date range exceeds the maximum limit of 31 days","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Maximum last 6 months",
			userClaims:     userClaims,
			query:          "startDate=2024-03-01T00:00:00%2b07:00&endDate=2024-03-03T23:59:59%2b07:00", // NOSONAR
			timezone:       constant.TimeLoc,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"date range exceeds the maximum limit of the last 6 months","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			query:      validDateQuery,
			timezone:   constant.TimeLoc,
			setupMock: func() {
				service.On(
					"GetMerchantTransactionHistoryList", mock.Anything, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			query:      validDateQuery,
			timezone:   constant.TimeLoc,
			setupMock: func() {
				service.On(
					"GetMerchantTransactionHistoryList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []walletTransactionModel.MerchantTransactionHistoryListResp{{}},
					Meta: commonModel.Meta{},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"id":"","referenceId":"","type":"","channel":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","amount":0,"status":"","settlementStatus":""}],"pagination":{"page":0,"perPage":0,"totalItems":0,"totalPages":0}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/transactions", nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			req.URL.RawQuery = test.query
			req.Header.Set(constant.HeaderTimezone, test.timezone)

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestExportMerchantTransactionHistoryList(t *testing.T) {
	validDateQuery := generateValidDateQuery()

	vld := validatorExt.New()
	service := serviceMock.NewIWalletTransactionService(t)

	handler := New(vld, service)

	route := chi.NewRouter()
	route.Get("/transactions/export", handler.ExportMerchantTransactionHistoryList)

	userClaims := &user.UserTokenClaims{
		MerchantId: "431fc9ae-4026-4190-a2bd-83c1ccf7cc6a", // NOSONAR
	}

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		query          string
		timezone       string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Invalid mandatory field",
			userClaims:     userClaims,
			query:          "startDate=2025-02-27T00:00:00%2b07:00&endDate=2025-02-03T23:59:59%2b07:00",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"EndDateReq","message":"Key: 'MerchantTransactionHistoryListReq.EndDateReq' Error:Field validation for 'EndDateReq' failed on the 'gtecsfield' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Time range exceeds the limit",
			userClaims:     userClaims,
			query:          "startDate=2025-01-01T00:00:00%2b07:00&endDate=2025-03-03T23:59:59%2b07:00", // NOSONAR
			timezone:       constant.TimeLoc,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"date range exceeds the maximum limit of 31 days","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Maximum last 6 months",
			userClaims:     userClaims,
			query:          "startDate=2024-03-01T00:00:00%2b07:00&endDate=2024-03-03T23:59:59%2b07:00", // NOSONAR
			timezone:       constant.TimeLoc,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"date range exceeds the maximum limit of the last 6 months","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			query:      validDateQuery,
			timezone:   constant.TimeLoc,
			setupMock: func() {
				service.On(
					"ExportMerchantTransactionHistoryList", mock.Anything, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			query:      validDateQuery,
			timezone:   constant.TimeLoc,
			setupMock: func() {
				service.On(
					"ExportMerchantTransactionHistoryList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.ExportResponse{
					DownloadURL: "https://", ExpiresAt: time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC),
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"downloadURL":"https://","expiresAt":"2025-03-01T12:00:00Z"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/transactions/export", nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			req.URL.RawQuery = test.query
			req.Header.Set(constant.HeaderTimezone, test.timezone)

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetMerchantTransactionDetail(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMock.NewIWalletTransactionService(t)

	handler := New(vld, service)

	route := chi.NewRouter()
	route.Get("/transactions/{id}", handler.GetMerchantTransactionDetail)

	userClaims := &user.UserTokenClaims{
		MerchantId: "431fc9ae-4026-4190-a2bd-83c1ccf7cc6a", // NOSONAR
	}
	trxId := "52bc968d-65b0-4d82-8d70-3c9d7d1fd6aa"

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		id             string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR:Invalid transaction id format",
			userClaims:     userClaims,
			id:             "ABC",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid transaction id format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"GetMerchantTransactionDetail", mock.Anything, mock.Anything, trxId,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"GetMerchantTransactionDetail", mock.Anything, mock.Anything, trxId,
				).Once().Return(&walletTransactionModel.MerchantTransactionDetailResp{
					Id:        trxId,
					Type:      "MERCHANT_TOP_UP", // NOSONAR
					Channel:   "VIRTUAL_ACCOUNT", // NOSONAR
					Amount:    50_000,
					CreatedBy: "Test",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"52bc968d-65b0-4d82-8d70-3c9d7d1fd6aa","referenceId":"","type":"MERCHANT_TOP_UP","channel":"VIRTUAL_ACCOUNT","createdAt":"","updatedAt":"","amount":50000,"createdBy":"Test","additionalInfo":null,"status":"","settlementStatus":""}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.id == "" {
				test.id = trxId
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/transactions/"+test.id, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
