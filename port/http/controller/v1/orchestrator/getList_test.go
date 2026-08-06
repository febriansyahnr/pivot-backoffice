package orchestratorController_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/orchestrator"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	reportingSvc := serviceMocks.NewIReportingService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	handler := New(
		&config.Config{AppConfig: config.AppConfig{PaginationPerPage: 1}}, orchestratorSvc, nil, nil, reportingSvc,
	)

	router := chi.NewRouter()
	router.Get("/transactions", handler.GetList)

	now := time.Now().UTC()
	validStartDateSimple := now.AddDate(0, 0, -30).Format("2006-01-02")
	validEndDateSimple := now.AddDate(0, 0, -29).Format("2006-01-02")

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		setupReq       func(r *http.Request) *http.Request
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Empty filter date",
			userClaims:     &user.UserTokenClaims{},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, "start settlement date and end settlement date must be filled"),
		},
		{
			name:       "ERROR:Invalid start date format",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=xxxx-xx-xx&endSettlementDate=%s", validStartDateSimple)
				return r
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, c.ErrInvalidStartDateFmt.Error()),
		},
		{
			name:       "ERROR:Invalid end date format",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=xxxx-xx-xx", validStartDateSimple)
				return r
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, c.ErrInvalidEndDateFmt.Error()),
		},
		{
			name:       "ERROR:Invalid filter field",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s&sort=xxx", validStartDateSimple, validStartDateSimple)
				return r
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, "invalid data sorting column"),
		},
		{
			name:       "ERROR:Invalid date input 2",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s", validEndDateSimple, validStartDateSimple)
				return r
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, c.ErrFilterDateInput.Error()),
		},
		{
			name:       "ERROR:Invalid page number format",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s&page=x", validStartDateSimple, validEndDateSimple)
				return r
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, response.ErrTypeAPI, "invalid page format. Use number format instead"),
		},
		{
			name:       "ERROR:Some error when call ListBalanceHistory function",
			userClaims: &user.UserTokenClaims{MerchantId: "79f52ad0-4820-46fd-8d38-10e5a0a8514c"},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s&page=1", validStartDateSimple, validEndDateSimple)
				return r
			},
			setupMock: func() {
				reportingSvc.On(
					"ListBalanceHistory", c.ValueCtxMockType(), mock.Anything, c.Int64MockType(), c.Int64MockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, response.ErrTypeAPI, "some error"),
		},
		{
			name:       "ERROR:Some error when call GetList function",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s&page=1", validStartDateSimple, validEndDateSimple)
				return r
			},
			setupMock: func() {
				orchestratorSvc.On(
					"GetList", c.ValueCtxMockType(), mock.Anything, c.Int64MockType(), c.Int64MockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, response.ErrTypeAPI, "some error"),
		},
		{
			name:       "SUCCESS",
			userClaims: &user.UserTokenClaims{},
			setupReq: func(r *http.Request) *http.Request {
				r.URL.RawQuery = fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s&page=1", validStartDateSimple, validEndDateSimple)
				return r
			},
			setupMock: func() {
				orchestratorSvc.On(
					"GetList", c.ValueCtxMockType(), mock.Anything, c.Int64MockType(), c.Int64MockType(),
				).Return(&commonModel.PaginationResponse{
					Data: map[string]string{"message": "OK"}, Meta: commonModel.Meta{},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK", "data":{"message": "OK"}, "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/transactions", nil)

			if test.setupReq != nil {
				req = test.setupReq(req)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())

			reportingSvc.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}

func TestGetOpenApiBalanceHistories(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)
	reportingSvc := serviceMocks.NewIReportingService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	route := chi.NewRouter()
	route.Get(
		"/transaction-histories", New(&config.Config{}, orchestratorSvc, merchantSvc, validator.New(), reportingSvc).GetOpenApiBalanceHistories,
	)

	now := time.Now().UTC()
	validStartDate := now.AddDate(0, 0, -30).Format(time.RFC3339)       // 30 days ago
	validEndDate := now.AddDate(0, 0, -29).Format(time.RFC3339)         // 29 days ago
	invalidRangeStartDate := now.AddDate(-1, 0, 0).Format(time.RFC3339) // 1 year ago
	invalidRangeEndDate := now.AddDate(0, 0, -1).Format(time.RFC3339)   // 1 day ago

	merchantInfo := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: "b0b298a9-69f8-49fa-88ae-bfa8802e1224",
	}
	subMerchantId := "cef1062b-aef9-4006-b35d-29ababce3de5"
	subMerchantHeader := map[string]string{c.HeaderXSubMerchantID: subMerchantId}
	respGeneralErr := `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`
	tests := []struct {
		name           string
		params         string
		claims         *merchantModel.MerchantAuthTokenClaims
		headers        map[string]string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Merchant auth token not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid sort order",
			claims:         merchantInfo,
			params:         "transactionType=VA_PAYMENT&accountType=PAYMENT&sort=XXXXX", // NOSONAR
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_value_invalid","message":"invalid sort order","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid sort order"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid start date format",
			claims:         merchantInfo,
			params:         "transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=XXX", // NOSONAR
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_value_invalid","message":"invalid start date format","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid start date format"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid end date format",
			claims:         merchantInfo,
			params:         fmt.Sprintf("transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=%s&endDate=123", validStartDate), // NOSONAR
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_value_invalid","message":"invalid end date format","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid end date format"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Filter start date input after end date",
			claims:         merchantInfo,
			params:         fmt.Sprintf("transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=%s&endDate=%s", validEndDate, validStartDate), // NOSONAR
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_value_invalid","message":"invalid filter date value","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid filter date value"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid date range input",
			claims:         merchantInfo,
			params:         fmt.Sprintf("transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=%s&endDate=%s", invalidRangeStartDate, invalidRangeEndDate), // NOSONAR
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_value_invalid","message":"date range exceeds limits","error":{"type":"API_ERROR","details":[{"field":"","message":"date range exceeds limits"}],"traceId":""}}`,
		},
		{
			name:    "ERROR:Find sub merchant by id",
			claims:  merchantInfo,
			headers: subMerchantHeader,                                                                                                              // NOSONAR
			params:  fmt.Sprintf("transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=%s&endDate=%s", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				merchantSvc.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   respGeneralErr, // NOSONAR
		},
		{
			name:    "ERROR:Access to merchant id is not allowed",
			claims:  merchantInfo,
			headers: subMerchantHeader,                                                                                                              // NOSONAR
			params:  fmt.Sprintf("transactionType=VA_PAYMENT&accountType=PAYMENT&sort=-date&startDate=%s&endDate=%s", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				merchantSvc.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(&merchantModel.Merchant{}, nil)
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"forbidden_access","message":"Provided API Key does not have the correct permissions to perform the operation","error":{"type":"API_ERROR","details":[{"field":"","message":"forbidden access"}],"traceId":""}}`,
		},
		{
			name:    "ERROR:Invalid page",
			claims:  merchantInfo,
			headers: subMerchantHeader,                                                            // NOSONAR
			params:  fmt.Sprintf("startDate=%s&endDate=%s&page=-1", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				merchantSvc.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(&merchantModel.Merchant{ParentID: sql.NullString{String: merchantInfo.MerchantId}}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   respGeneralErr,
		},
		{
			name:    "ERROR:Invalid per page",
			claims:  merchantInfo,
			headers: subMerchantHeader,                                                                      // NOSONAR
			params:  fmt.Sprintf("startDate=%s&endDate=%s&page=1&perPage=-1", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				merchantSvc.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(&merchantModel.Merchant{ParentID: sql.NullString{String: merchantInfo.MerchantId}}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   respGeneralErr,
		},
		{
			name: "ERROR:List balance history",
			claims: &merchantModel.MerchantAuthTokenClaims{
				MerchantId: "79f52ad0-4820-46fd-8d38-10e5a0a8514c",
			},
			params: fmt.Sprintf("startDate=%s&endDate=%s&page=2&perPage=20", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				reportingSvc.On("ListBalanceHistory", mock.Anything, mock.MatchedBy(func(p *orchestratorModel.TransactionHistoryFilterRequest) bool {
					return p.StartSettlementDate.Format(time.RFC3339) == validStartDate &&
						p.EndSettlementDate.Format(time.RFC3339) == validEndDate
				}), int64(2), int64(20)).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   respGeneralErr, // NOSONAR
		},
		{
			name:   "ERROR:Get list",
			claims: merchantInfo,
			params: fmt.Sprintf("startDate=%s&endDate=%s&page=2&perPage=20", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				orchestratorSvc.On("GetList", mock.Anything, mock.MatchedBy(func(p *orchestratorModel.TransactionHistoryFilterRequest) bool {
					return p.StartSettlementDate.Format(time.RFC3339) == validStartDate &&
						p.EndSettlementDate.Format(time.RFC3339) == validEndDate
				}), int64(2), int64(20)).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   respGeneralErr, // NOSONAR
		},
		{
			name:   "SUCCESS",
			claims: merchantInfo,
			params: fmt.Sprintf("startDate=%s&endDate=%s&transactionId=REF0001&page=2&perPage=20", validStartDate, validEndDate), // NOSONAR
			setupMock: func() {
				orchestratorSvc.On(
					"GetList", mock.Anything, mock.MatchedBy(func(p *orchestratorModel.TransactionHistoryFilterRequest) bool {
						return p.StartSettlementDate.Format(time.RFC3339) == validStartDate &&
							p.EndSettlementDate.Format(time.RFC3339) == validEndDate &&
							p.TransactionId == "REF0001"
					}), int64(2), int64(20),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []any{},
					Meta: commonModel.Meta{
						Page:    2,
						PerPage: 20,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":[],"pagination":{"page":2,"perPage":20,"totalItems":0,"totalPages":0}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			url := "/transaction-histories"
			if test.params != "" {
				url += "?" + test.params
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			for key, val := range test.headers {
				req.Header.Set(key, val)
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, test.claims))
			}

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}

			merchantSvc.AssertExpectations(t)
			reportingSvc.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}
